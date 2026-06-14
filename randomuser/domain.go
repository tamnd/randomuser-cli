package randomuser

import (
	"context"
	"unicode"

	"github.com/tamnd/any-cli/kit"
	"github.com/tamnd/any-cli/kit/errs"
)

// domain.go exposes randomuser as a kit Domain driver.
//
// A multi-domain host (ant) enables it with a single blank import:
//
//	import _ "github.com/tamnd/randomuser-cli/randomuser"
//
// The same Domain also builds the standalone randomuser binary (see cli.NewApp).
func init() { kit.Register(Domain{}) }

// Domain is the randomuser driver.
type Domain struct{}

// Info describes the scheme, the hostnames a pasted link is matched against,
// and the identity reused for the binary's help and version.
func (Domain) Info() kit.DomainInfo {
	return kit.DomainInfo{
		Scheme: "randomuser",
		Hosts:  []string{Host},
		Identity: kit.Identity{
			Binary: "randomuser",
			Short:  "Generate random user profiles from randomuser.me",
			Long: `randomuser fetches randomly generated user profiles from
the public randomuser.me API. No login required.`,
			Site: Host,
			Repo: "https://github.com/tamnd/randomuser-cli",
		},
	}
}

// Register installs the client factory and every operation onto app.
func (Domain) Register(app *kit.App) {
	app.SetClient(newClient)

	// generate: fetch randomly generated user profiles
	kit.Handle(app, kit.OpMeta{
		Name:    "generate",
		Group:   "read",
		List:    true,
		Summary: "Generate random user profiles",
	}, generateOp)
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

type generateInput struct {
	Count  int     `kit:"flag" help:"number of users to generate" default:"1"`
	Nat    string  `kit:"flag" help:"nationality codes e.g. us,gb,au (comma-sep)"`
	Gender string  `kit:"flag" help:"gender filter: male|female"`
	Seed   string  `kit:"flag" help:"seed for reproducible results"`
	Client *Client `kit:"inject"`
}

// --- handlers ---

func generateOp(ctx context.Context, in generateInput, emit func(User) error) error {
	count := in.Count
	if count <= 0 {
		count = 1
	}
	items, err := in.Client.Generate(ctx, GenerateParams{
		Results: count,
		Nat:     in.Nat,
		Gender:  in.Gender,
		Seed:    in.Seed,
	})
	if err != nil {
		return err
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
// Numeric input is classified as "count"; otherwise "nationality".
func (Domain) Classify(input string) (uriType, id string, err error) {
	if input == "" {
		return "", "", errs.Usage("empty randomuser reference")
	}
	if isNumeric(input) {
		return "count", input, nil
	}
	return "nationality", input, nil
}

// Locate returns the live https URL for a (type, id).
func (Domain) Locate(uriType, id string) (string, error) {
	switch uriType {
	case "count", "nationality":
		return "https://randomuser.me", nil
	default:
		return "", errs.Usage("randomuser has no resource type %q", uriType)
	}
}

// isNumeric reports whether s consists entirely of digits.
func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if !unicode.IsDigit(r) {
			return false
		}
	}
	return true
}
