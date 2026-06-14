package catapi

import (
	"context"
	"time"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes catapi as a kit Domain driver.
//
// A multi-domain host (ant) enables it with a single blank import:
//
//	import _ "github.com/tamnd/catapi-cli/catapi"
//
// The same Domain also builds the standalone catapi binary (see cli.NewApp).
func init() { kit.Register(Domain{}) }

// Domain is the catapi driver.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "catapi",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "catapi",
			Short:  "Cat breed info and random images from TheCatAPI",
			Long: `catapi fetches cat breed information and random cat images from
TheCatAPI (api.thecatapi.com). No API key required for public endpoints.
Supports breed listing, name search, image search, and category listing.`,
			Site: Host,
			Repo: "https://github.com/tamnd/catapi-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	// search: find cat images, optionally filtered by breed
	kit.Handle(app, kit.OpMeta{
		Name:    "search",
		Group:   "read",
		List:    true,
		Summary: "Search for cat images",
	}, searchOp)

	// breeds: list all cat breeds or search by name
	kit.Handle(app, kit.OpMeta{
		Name:    "breeds",
		Group:   "read",
		List:    true,
		Summary: "List or search cat breeds",
	}, breedsOp)

	// categories: list image categories
	kit.Handle(app, kit.OpMeta{
		Name:    "categories",
		Group:   "read",
		List:    true,
		Summary: "List image categories",
	}, categoriesOp)
}

// newClient builds the client from host-resolved config.
func newClient(_ context.Context, cfg kit.Config) (any, error) {
	c := DefaultConfig()
	if cfg.UserAgent != "" {
		c.UserAgent = cfg.UserAgent
	}
	if cfg.Rate > 0 {
		c.Rate = cfg.Rate
	}
	if cfg.Retries > 0 {
		c.Retries = cfg.Retries
	}
	if cfg.Timeout > 0 {
		c.Timeout = cfg.Timeout
	}
	return NewClient(c), nil
}

// --- inputs ---

type searchInput struct {
	Limit     int           `kit:"flag,inherit" help:"max images to return"`
	Breed     string        `kit:"flag"         help:"filter by breed ID (e.g. beng)"`
	HasBreeds bool          `kit:"flag"         help:"only return images with breed info"`
	Delay     time.Duration `kit:"flag,inherit" help:"minimum spacing between requests"`
	Client    *Client       `kit:"inject"`
}

type breedsInput struct {
	Query  string        `kit:"flag"         help:"search term; if empty, list all breeds"`
	Limit  int           `kit:"flag,inherit" help:"max results"`
	Page   int           `kit:"flag"         help:"page number for breed listing (default 0)"`
	Delay  time.Duration `kit:"flag,inherit" help:"minimum spacing between requests"`
	Client *Client       `kit:"inject"`
}

type categoriesInput struct {
	Delay  time.Duration `kit:"flag,inherit" help:"minimum spacing between requests"`
	Client *Client       `kit:"inject"`
}

// --- handlers ---

func searchOp(ctx context.Context, in searchInput, emit func(Image) error) error {
	items, err := in.Client.Search(ctx, in.Limit, in.Breed, in.HasBreeds)
	if err != nil {
		return mapErr(err)
	}
	for _, item := range items {
		if err := emit(item); err != nil {
			return err
		}
	}
	return nil
}

func breedsOp(ctx context.Context, in breedsInput, emit func(Breed) error) error {
	var (
		items []Breed
		err   error
	)
	if in.Query != "" {
		items, err = in.Client.SearchBreeds(ctx, in.Query)
	} else {
		items, err = in.Client.Breeds(ctx, in.Limit, in.Page)
	}
	if err != nil {
		return mapErr(err)
	}
	for _, item := range items {
		if err := emit(item); err != nil {
			return err
		}
	}
	return nil
}

func categoriesOp(ctx context.Context, in categoriesInput, emit func(Category) error) error {
	items, err := in.Client.Categories(ctx)
	if err != nil {
		return mapErr(err)
	}
	for _, item := range items {
		if err := emit(item); err != nil {
			return err
		}
	}
	return nil
}

// --- Resolver: pure string functions, no network ---

// Classify turns an input into the canonical (type, id).
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("empty catapi reference")
	}
	return "breed", input, nil
}

// Locate returns the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "breed":
		return "https://api.thecatapi.com/v1/breeds/" + id, nil
	default:
		return "", errs.Usage("catapi has no resource type %q", uriType)
	}
}

// mapErr converts a library error into the kit error kind.
func mapErr(err error) error {
	return err
}
