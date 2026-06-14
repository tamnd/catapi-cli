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

const fakeBreedsJSON = `[{"id":"abys","name":"Abyssinian","origin":"Egypt","temperament":"Active, Energetic, Independent","life_span":"14 - 15","description":"The Abyssinian is easy to care for."},{"id":"siam","name":"Siamese","origin":"Thailand","temperament":"Active, Agile, Clever","life_span":"15 - 20","description":"The Siamese is a very old breed."}]`

const fakeSiameseJSON = `[{"id":"siam","name":"Siamese","origin":"Thailand","temperament":"Active, Agile, Clever","life_span":"15 - 20","description":"The Siamese is a very old breed."}]`

const fakeImagesJSON = `[{"id":"3v7","url":"https://cdn2.thecatapi.com/images/3v7.gif","width":500,"height":375,"breeds":[]},{"id":"6ne","url":"https://cdn2.thecatapi.com/images/6ne.jpg","width":1200,"height":800,"breeds":[{"id":"beng","name":"Bengal","temperament":"Alert, Agile","origin":"United States","life_span":"12 - 16"}]}]`

const fakeCategoriesJSON = `[{"id":1,"name":"hats"},{"id":2,"name":"space"},{"id":4,"name":"sunglasses"}]`

// --- helpers ---

func newTestClient(ts *httptest.Server) *catapi.Client {
	cfg := catapi.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return catapi.NewClient(cfg)
}

// --- Search tests ---

func TestSearchSendsUserAgent(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = fmt.Fprint(w, fakeImagesJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Search(context.Background(), 2, "", false)
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent header not sent")
	}
}

func TestSearchParsesImages(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeImagesJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Search(context.Background(), 2, "", false)
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
		t.Errorf("items[0].URL = %q, want cdn2.thecatapi.com", first.URL)
	}
	if first.Width != 500 {
		t.Errorf("items[0].Width = %d, want 500", first.Width)
	}
	// first image has no breeds
	if first.Breed != "" {
		t.Errorf("items[0].Breed = %q, want empty", first.Breed)
	}

	// second image has Bengal breed info
	second := items[1]
	if second.Breed != "Bengal" {
		t.Errorf("items[1].Breed = %q, want Bengal", second.Breed)
	}
	if second.Origin != "United States" {
		t.Errorf("items[1].Origin = %q, want United States", second.Origin)
	}
}

func TestSearchPassesBreedParam(t *testing.T) {
	var gotQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = fmt.Fprint(w, fakeImagesJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Search(context.Background(), 3, "beng", false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotQuery, "breed_ids=beng") {
		t.Errorf("query = %q, want contains breed_ids=beng", gotQuery)
	}
}

func TestSearchHasBreedsParam(t *testing.T) {
	var gotQuery string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.RawQuery
		_, _ = fmt.Fprint(w, fakeImagesJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Search(context.Background(), 5, "", true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotQuery, "has_breeds=1") {
		t.Errorf("query = %q, want contains has_breeds=1", gotQuery)
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
}

// --- Categories tests ---

func TestCategoriesParsesItems(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeCategoriesJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Categories(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 3 {
		t.Fatalf("len(items) = %d, want 3", len(items))
	}
	if items[0].ID != 1 || items[0].Name != "hats" {
		t.Errorf("items[0] = %+v, want {ID:1 Name:hats}", items[0])
	}
	if items[2].Name != "sunglasses" {
		t.Errorf("items[2].Name = %q, want sunglasses", items[2].Name)
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
