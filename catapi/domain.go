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
Supports breed listing, name search, and image retrieval.`,
			Site: Host,
			Repo: "https://github.com/tamnd/catapi-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	// breeds: list all cat breeds (paginated)
	kit.Handle(app, kit.OpMeta{
		Name:    "breeds",
		Group:   "read",
		List:    true,
		Summary: "List all cat breeds",
	}, breedsOp)

	// search: search breeds by name
	kit.Handle(app, kit.OpMeta{
		Name:    "search",
		Group:   "read",
		List:    true,
		Summary: "Search cat breeds by name",
	}, searchOp)

	// images: get random cat images
	kit.Handle(app, kit.OpMeta{
		Name:    "images",
		Group:   "read",
		List:    true,
		Summary: "Get random cat images",
	}, imagesOp)
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

type breedsInput struct {
	Limit  int           `kit:"flag,inherit" help:"max breeds to return (default 25)"`
	Page   int           `kit:"flag"         help:"page number for breed listing (default 0)"`
	Delay  time.Duration `kit:"flag,inherit" help:"minimum spacing between requests"`
	Client *Client       `kit:"inject"`
}

type searchInput struct {
	Query  string        `kit:"arg"          help:"breed name to search for"`
	Delay  time.Duration `kit:"flag,inherit" help:"minimum spacing between requests"`
	Client *Client       `kit:"inject"`
}

type imagesInput struct {
	Limit  int           `kit:"flag,inherit" help:"max images to return (default 5)"`
	Delay  time.Duration `kit:"flag,inherit" help:"minimum spacing between requests"`
	Client *Client       `kit:"inject"`
}

// --- handlers ---

func breedsOp(ctx context.Context, in breedsInput, emit func(Breed) error) error {
	limit := in.Limit
	if limit <= 0 {
		limit = 25
	}
	items, err := in.Client.Breeds(ctx, limit, in.Page)
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

func searchOp(ctx context.Context, in searchInput, emit func(Breed) error) error {
	items, err := in.Client.SearchBreeds(ctx, in.Query)
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

func imagesOp(ctx context.Context, in imagesInput, emit func(CatImage) error) error {
	limit := in.Limit
	if limit <= 0 {
		limit = 5
	}
	items, err := in.Client.Images(ctx, limit)
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
