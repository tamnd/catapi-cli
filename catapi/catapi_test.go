package catapi_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tamnd/catapi-cli/catapi"
)

// --- fixture JSON ---

const fakeBreedsJSON = `[{"id":"abys","name":"Abyssinian","origin":"Egypt","temperament":"Active, Energetic, Independent","life_span":"14 - 15","weight":{"imperial":"7 - 10","metric":"3 - 5"},"intelligence":5,"affection_level":5,"energy_level":4,"child_friendly":3,"wikipedia_url":"https://en.wikipedia.org/wiki/Abyssinian_cat"},{"id":"siam","name":"Siamese","origin":"Thailand","temperament":"Active, Agile, Clever","life_span":"15 - 20","weight":{"imperial":"8 - 15","metric":"4 - 7"},"intelligence":5,"affection_level":5,"energy_level":4,"child_friendly":4,"wikipedia_url":"https://en.wikipedia.org/wiki/Siamese_cat"}]`

const fakeSiameseJSON = `[{"id":"siam","name":"Siamese","origin":"Thailand","temperament":"Active, Agile, Clever","life_span":"15 - 20","weight":{"imperial":"8 - 15","metric":"4 - 7"},"intelligence":5,"affection_level":5,"energy_level":4,"child_friendly":4,"wikipedia_url":"https://en.wikipedia.org/wiki/Siamese_cat"}]`

const fakeImagesJSON = `[{"id":"3v7","url":"https://cdn2.thecatapi.com/images/3v7.gif","width":500,"height":375},{"id":"6ne","url":"https://cdn2.thecatapi.com/images/6ne.jpg","width":1200,"height":800}]`

// --- helpers ---

func newTestClient(ts *httptest.Server) *catapi.Client {
	cfg := catapi.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return catapi.NewClient(cfg)
}

// --- Images tests ---

func TestImagesSendsUserAgent(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = fmt.Fprint(w, fakeImagesJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Images(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent header not sent")
	}
}

func TestImagesParsesItems(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeImagesJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Images(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}

	first := items[0]
	if first.ID != "3v7" {
		t.Errorf("items[0].ID = %q, want 3v7", first.ID)
	}
	if !strings.Contains(first.URL, "cdn2.thecatapi.com") {
		t.Errorf("items[0].URL = %q, want cdn2.thecatapi.com URL", first.URL)
	}
	if first.Width != 500 {
		t.Errorf("items[0].Width = %d, want 500", first.Width)
	}
	if first.Height != 375 {
		t.Errorf("items[0].Height = %d, want 375", first.Height)
	}

	second := items[1]
	if second.ID != "6ne" {
		t.Errorf("items[1].ID = %q, want 6ne", second.ID)
	}
	if second.Width != 1200 {
		t.Errorf("items[1].Width = %d, want 1200", second.Width)
	}
}

func TestImagesLimitParam(t *testing.T) {
	var gotQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = fmt.Fprint(w, fakeImagesJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Images(context.Background(), 3)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotQuery, "limit=3") {
		t.Errorf("query = %q, want contains limit=3", gotQuery)
	}
}

func TestImagesDefaultLimit(t *testing.T) {
	var gotQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = fmt.Fprint(w, fakeImagesJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Images(context.Background(), 0)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotQuery, "limit=5") {
		t.Errorf("query = %q, want contains limit=5 (default)", gotQuery)
	}
}

// --- Breeds tests ---

func TestBreedsParsesItems(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeBreedsJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Breeds(context.Background(), 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}
	got := items[0]
	if got.ID != "abys" {
		t.Errorf("items[0].ID = %q, want abys", got.ID)
	}
	if got.Name != "Abyssinian" {
		t.Errorf("items[0].Name = %q, want Abyssinian", got.Name)
	}
	if got.Origin != "Egypt" {
		t.Errorf("items[0].Origin = %q, want Egypt", got.Origin)
	}
	if !strings.Contains(got.Temperament, "Active") {
		t.Errorf("items[0].Temperament = %q, want contains Active", got.Temperament)
	}
	if got.WeightMetric != "3 - 5" {
		t.Errorf("items[0].WeightMetric = %q, want 3 - 5", got.WeightMetric)
	}
	if got.Intelligence != 5 {
		t.Errorf("items[0].Intelligence = %d, want 5", got.Intelligence)
	}
	if got.AffectionLevel != 5 {
		t.Errorf("items[0].AffectionLevel = %d, want 5", got.AffectionLevel)
	}
	if got.EnergyLevel != 4 {
		t.Errorf("items[0].EnergyLevel = %d, want 4", got.EnergyLevel)
	}
	if got.ChildFriendly != 3 {
		t.Errorf("items[0].ChildFriendly = %d, want 3", got.ChildFriendly)
	}
	if !strings.Contains(got.WikipediaURL, "wikipedia.org") {
		t.Errorf("items[0].WikipediaURL = %q, want contains wikipedia.org", got.WikipediaURL)
	}
	if items[1].Name != "Siamese" {
		t.Errorf("items[1].Name = %q, want Siamese", items[1].Name)
	}
}

func TestSearchBreedsQuery(t *testing.T) {
	var gotPath string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path + "?" + r.URL.RawQuery
		_, _ = fmt.Fprint(w, fakeSiameseJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.SearchBreeds(context.Background(), "siamese")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotPath, "/v1/breeds/search") {
		t.Errorf("path = %q, want /v1/breeds/search", gotPath)
	}
	if !strings.Contains(gotPath, "q=siamese") {
		t.Errorf("path = %q, want contains q=siamese", gotPath)
	}
	if len(items) != 1 || items[0].Name != "Siamese" {
		t.Errorf("items = %v, want one Siamese", items)
	}
	if items[0].WeightMetric != "4 - 7" {
		t.Errorf("items[0].WeightMetric = %q, want 4 - 7", items[0].WeightMetric)
	}
}

// --- Retry test ---

func TestBreedsRetriesOn503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, fakeBreedsJSON)
	}))
	defer ts.Close()

	cfg := catapi.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 3
	c := catapi.NewClient(cfg)

	_, err := c.Breeds(context.Background(), 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}
