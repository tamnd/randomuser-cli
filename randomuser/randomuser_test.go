package randomuser_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/tamnd/randomuser-cli/randomuser"
)

// fakeMockJSON has two users: one US female, one FR male.
// login.uuid and login.username are included per the inc parameter.
const fakeMockJSON = `{"results":[{"gender":"female","name":{"title":"Ms","first":"Cassandra","last":"Olson"},"location":{"street":{"number":123,"name":"Main St"},"city":"Portland","state":"Oregon","country":"United States","postcode":"97201","coordinates":{"latitude":"45.5231","longitude":"-122.6765"}},"email":"cassandra.olson@example.com","login":{"uuid":"abc-def-001","username":"goldenmouse123","password":"chester"},"dob":{"date":"1985-02-14T00:00:00.000Z","age":41},"phone":"(503) 555-0142","nat":"US"},{"gender":"male","name":{"title":"Mr","first":"Jean","last":"Dupont"},"location":{"street":{"number":7,"name":"Rue de Rivoli"},"city":"Paris","state":"Ile-de-France","country":"France","postcode":75001,"coordinates":{"latitude":"48.8566","longitude":"2.3522"}},"email":"jean.dupont@example.com","login":{"uuid":"xyz-ghi-002","username":"bluefox456","password":"secret"},"dob":{"date":"1992-07-22T00:00:00.000Z","age":33},"phone":"01-234-567","nat":"FR"}],"info":{"seed":"abc123","results":2,"page":1,"version":"1.4"}}`

func newTestClient(ts *httptest.Server) *randomuser.Client {
	cfg := randomuser.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	return randomuser.NewClient(cfg)
}

func TestGenerateSendsUserAgent(t *testing.T) {
	var gotUA string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		_, _ = fmt.Fprint(w, fakeMockJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Generate(context.Background(), randomuser.GenerateParams{Results: 2})
	if err != nil {
		t.Fatal(err)
	}
	if gotUA == "" {
		t.Error("User-Agent not sent")
	}
}

func TestGenerateParsesUsers(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, fakeMockJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	items, err := c.Generate(context.Background(), randomuser.GenerateParams{Results: 2})
	if err != nil {
		t.Fatal(err)
	}
	if len(items) != 2 {
		t.Fatalf("len(items) = %d, want 2", len(items))
	}

	u := items[0]
	if u.UUID != "abc-def-001" {
		t.Errorf("UUID = %q, want abc-def-001", u.UUID)
	}
	if u.Name != "Cassandra Olson" {
		t.Errorf("Name = %q, want Cassandra Olson", u.Name)
	}
	if u.Email != "cassandra.olson@example.com" {
		t.Errorf("Email = %q, want cassandra.olson@example.com", u.Email)
	}
	if u.City != "Portland" {
		t.Errorf("City = %q, want Portland", u.City)
	}
	if u.State != "Oregon" {
		t.Errorf("State = %q, want Oregon", u.State)
	}
	if u.Age != 41 {
		t.Errorf("Age = %d, want 41", u.Age)
	}
	if u.Gender != "female" {
		t.Errorf("Gender = %q, want female", u.Gender)
	}
	if u.Username != "goldenmouse123" {
		t.Errorf("Username = %q, want goldenmouse123", u.Username)
	}
	if u.Nat != "US" {
		t.Errorf("Nat = %q, want US", u.Nat)
	}

	u2 := items[1]
	if u2.UUID != "xyz-ghi-002" {
		t.Errorf("items[1].UUID = %q, want xyz-ghi-002", u2.UUID)
	}
	if u2.Name != "Jean Dupont" {
		t.Errorf("items[1].Name = %q, want Jean Dupont", u2.Name)
	}
	if u2.Nat != "FR" {
		t.Errorf("items[1].Nat = %q, want FR", u2.Nat)
	}
	if u2.Username != "bluefox456" {
		t.Errorf("items[1].Username = %q, want bluefox456", u2.Username)
	}
}

func TestGenerateWithNatPassedInURL(t *testing.T) {
	var gotURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		_, _ = fmt.Fprint(w, fakeMockJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Generate(context.Background(), randomuser.GenerateParams{Results: 1, Nat: "us"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotURL, "nat=us") {
		t.Errorf("URL %q does not contain nat=us", gotURL)
	}
}

func TestGenerateWithGenderPassedInURL(t *testing.T) {
	var gotURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		_, _ = fmt.Fprint(w, fakeMockJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Generate(context.Background(), randomuser.GenerateParams{Results: 1, Gender: "female"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotURL, "gender=female") {
		t.Errorf("URL %q does not contain gender=female", gotURL)
	}
}

func TestGenerateWithSeedPassedInURL(t *testing.T) {
	var gotURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		_, _ = fmt.Fprint(w, fakeMockJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Generate(context.Background(), randomuser.GenerateParams{Results: 2, Seed: "foobar"})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotURL, "seed=foobar") {
		t.Errorf("URL %q does not contain seed=foobar", gotURL)
	}
}

func TestGenerateRetriesOn503(t *testing.T) {
	var hits int
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		if hits < 3 {
			w.WriteHeader(http.StatusServiceUnavailable)
			return
		}
		_, _ = fmt.Fprint(w, fakeMockJSON)
	}))
	defer ts.Close()

	cfg := randomuser.DefaultConfig()
	cfg.BaseURL = ts.URL
	cfg.Rate = 0
	cfg.Retries = 3
	c := randomuser.NewClient(cfg)

	_, err := c.Generate(context.Background(), randomuser.GenerateParams{Results: 2})
	if err != nil {
		t.Fatal(err)
	}
	if hits != 3 {
		t.Errorf("server saw %d hits, want 3", hits)
	}
}

func TestGenerateIncludesLoginInIncParam(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		inc := r.URL.Query().Get("inc")
		if !strings.Contains(inc, "login") {
			t.Errorf("inc param %q does not include login", inc)
		}
		_, _ = fmt.Fprint(w, fakeMockJSON)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Generate(context.Background(), randomuser.GenerateParams{Results: 1})
	if err != nil {
		t.Fatal(err)
	}
}

func TestGenerateDefaultResultsIsOne(t *testing.T) {
	var gotURL string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotURL = r.URL.String()
		_, _ = fmt.Fprint(w, `{"results":[{"gender":"male","name":{"title":"Mr","first":"Bob","last":"Smith"},"location":{"city":"NY","state":"NY","country":"US","postcode":"10001","coordinates":{"latitude":"0","longitude":"0"}},"email":"bob@example.com","login":{"uuid":"u1","username":"bob1","password":"pw"},"dob":{"date":"1990-01-01T00:00:00.000Z","age":34},"phone":"555-0000","nat":"US"}],"info":{"seed":"x","results":1,"page":1,"version":"1.4"}}`)
	}))
	defer ts.Close()

	c := newTestClient(ts)
	_, err := c.Generate(context.Background(), randomuser.GenerateParams{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(gotURL, "results=1") {
		t.Errorf("URL %q does not contain results=1 (default should be 1)", gotURL)
	}
}
