// Package catapi is the library behind the catapi command line:
// the HTTP client, request shaping, and typed data models for TheCatAPI.
//
// The free tier at api.thecatapi.com requires no API key. The Client sets a
// polite User-Agent, paces requests, and retries transient failures (429 and
// 5xx) with a capped backoff.
package catapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	neturl "net/url"
	"sync"
	"time"
)

// Host is the site this client talks to.
const Host = "api.thecatapi.com"

// Config holds all tunable parameters for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns a Config with sensible defaults for the free API tier.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://api.thecatapi.com",
		UserAgent: "catapi-cli/0.1 (tamnd87@gmail.com)",
		Rate:      500 * time.Millisecond,
		Timeout:   15 * time.Second,
		Retries:   3,
	}
}

// Client talks to TheCatAPI over HTTP.
type Client struct {
	cfg  Config
	http *http.Client
	mu   sync.Mutex
	last time.Time
}

// NewClient returns a Client configured with cfg.
func NewClient(cfg Config) *Client {
	return &Client{
		cfg:  cfg,
		http: &http.Client{Timeout: cfg.Timeout},
	}
}

// Images returns random cat images. limit controls the maximum number of
// results (default 5 if <= 0). Calls GET /v1/images/search.
func (c *Client) Images(ctx context.Context, limit int) ([]CatImage, error) {
	if limit <= 0 {
		limit = 5
	}
	params := neturl.Values{}
	params.Set("limit", fmt.Sprintf("%d", limit))
	u := fmt.Sprintf("%s/v1/images/search?%s", c.cfg.BaseURL, params.Encode())
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw []wireImage
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode images: %w", err)
	}
	out := make([]CatImage, 0, len(raw))
	for _, r := range raw {
		out = append(out, CatImage{
			ID:     r.ID,
			URL:    r.URL,
			Width:  r.Width,
			Height: r.Height,
		})
	}
	return out, nil
}

// Breeds returns all cat breeds paginated. limit and page are passed directly
// to the API. Calls GET /v1/breeds.
func (c *Client) Breeds(ctx context.Context, limit, page int) ([]Breed, error) {
	params := neturl.Values{}
	if limit > 0 {
		params.Set("limit", fmt.Sprintf("%d", limit))
	}
	params.Set("page", fmt.Sprintf("%d", page))
	u := fmt.Sprintf("%s/v1/breeds?%s", c.cfg.BaseURL, params.Encode())
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw []wireBreed
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode breeds: %w", err)
	}
	out := make([]Breed, 0, len(raw))
	for _, r := range raw {
		out = append(out, breedFromWire(r))
	}
	return out, nil
}

// SearchBreeds searches breeds by name query. Calls GET /v1/breeds/search.
func (c *Client) SearchBreeds(ctx context.Context, query string) ([]Breed, error) {
	u := fmt.Sprintf("%s/v1/breeds/search?q=%s", c.cfg.BaseURL, neturl.QueryEscape(query))
	body, err := c.get(ctx, u)
	if err != nil {
		return nil, err
	}
	var raw []wireBreed
	if err := json.Unmarshal(body, &raw); err != nil {
		return nil, fmt.Errorf("decode breed search: %w", err)
	}
	out := make([]Breed, 0, len(raw))
	for _, r := range raw {
		out = append(out, breedFromWire(r))
	}
	return out, nil
}

func breedFromWire(r wireBreed) Breed {
	return Breed{
		ID:             r.ID,
		Name:           r.Name,
		Origin:         r.Origin,
		Temperament:    r.Temperament,
		LifeSpan:       r.LifeSpan,
		WeightMetric:   r.Weight.Metric,
		Intelligence:   r.Intelligence,
		AffectionLevel: r.AffectionLevel,
		EnergyLevel:    r.EnergyLevel,
		ChildFriendly:  r.ChildFriendly,
		WikipediaURL:   r.WikipediaURL,
	}
}

func (c *Client) get(ctx context.Context, url string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, url)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", url, lastErr)
}

func (c *Client) do(ctx context.Context, rawURL string) ([]byte, bool, error) {
	c.pace()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, false, err
	}
	req.Header.Set("User-Agent", c.cfg.UserAgent)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, true, err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode == http.StatusTooManyRequests || resp.StatusCode >= 500 {
		return nil, true, fmt.Errorf("http %d", resp.StatusCode)
	}
	if resp.StatusCode != http.StatusOK {
		return nil, false, fmt.Errorf("http %d", resp.StatusCode)
	}
	b, err := io.ReadAll(resp.Body)
	return b, err != nil, err
}

func (c *Client) pace() {
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.cfg.Rate <= 0 {
		return
	}
	if wait := c.cfg.Rate - time.Since(c.last); wait > 0 {
		time.Sleep(wait)
	}
	c.last = time.Now()
}

func backoff(attempt int) time.Duration {
	d := time.Duration(attempt) * 500 * time.Millisecond
	if d > 5*time.Second {
		return 5 * time.Second
	}
	return d
}
