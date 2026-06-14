// Package randomuser is the library behind the randomuser command line:
// the HTTP client, request shaping, and the typed data models for randomuser.me.
//
// The Client sets a real User-Agent, paces requests so a busy session stays
// polite, and retries the transient failures (429 and 5xx) that any public
// site throws under load.
package randomuser

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

// Host is the site this client talks to.
const Host = "randomuser.me"

// Config holds all tunable parameters for the Client.
type Config struct {
	BaseURL   string
	UserAgent string
	Rate      time.Duration
	Timeout   time.Duration
	Retries   int
}

// DefaultConfig returns a Config with sensible defaults.
func DefaultConfig() Config {
	return Config{
		BaseURL:   "https://randomuser.me",
		UserAgent: "Mozilla/5.0 (compatible; randomuser-cli/dev; +https://github.com/tamnd/randomuser-cli)",
		Rate:      500 * time.Millisecond,
		Timeout:   15 * time.Second,
		Retries:   3,
	}
}

// Client talks to randomuser.me over HTTP.
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

// GenerateParams holds all parameters for the generate API call.
type GenerateParams struct {
	Results int
	Nat     string
	Gender  string
	Seed    string
}

// Generate fetches randomly generated user profiles with the given params.
func (c *Client) Generate(ctx context.Context, p GenerateParams) ([]User, error) {
	if p.Results <= 0 {
		p.Results = 5
	}
	if p.Results > 5000 {
		p.Results = 5000
	}

	q := url.Values{}
	q.Set("results", fmt.Sprintf("%d", p.Results))
	q.Set("inc", "name,email,location,phone,nat,gender,dob")
	if p.Nat != "" {
		q.Set("nat", p.Nat)
	}
	if p.Gender != "" {
		q.Set("gender", p.Gender)
	}
	if p.Seed != "" {
		q.Set("seed", p.Seed)
	}

	rawURL := c.cfg.BaseURL + "/api/?" + q.Encode()
	body, err := c.get(ctx, rawURL)
	if err != nil {
		return nil, err
	}

	var resp apiResponse
	if err := json.Unmarshal(body, &resp); err != nil {
		return nil, fmt.Errorf("decode users: %w", err)
	}

	users := make([]User, 0, len(resp.Results))
	for _, r := range resp.Results {
		dob := r.DOB.Date
		if len(dob) >= 10 {
			dob = dob[:10]
		}
		users = append(users, User{
			Gender:    r.Gender,
			Title:     r.Name.Title,
			FirstName: r.Name.First,
			LastName:  r.Name.Last,
			Email:     r.Email,
			Phone:     r.Phone,
			Nat:       r.Nat,
			City:      r.Location.City,
			State:     r.Location.State,
			Country:   r.Location.Country,
			Postcode:  parsePostcode(r.Location.Postcode),
			Latitude:  r.Location.Coordinates.Latitude,
			Longitude: r.Location.Coordinates.Longitude,
			BirthDate: dob,
			Age:       r.DOB.Age,
		})
	}
	return users, nil
}

func (c *Client) get(ctx context.Context, rawURL string) ([]byte, error) {
	var lastErr error
	for attempt := 0; attempt <= c.cfg.Retries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-time.After(backoff(attempt)):
			}
		}
		body, retry, err := c.do(ctx, rawURL)
		if err == nil {
			return body, nil
		}
		lastErr = err
		if !retry {
			return nil, err
		}
	}
	return nil, fmt.Errorf("get %s: %w", rawURL, lastErr)
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
	return min(time.Duration(attempt)*500*time.Millisecond, 5*time.Second)
}
