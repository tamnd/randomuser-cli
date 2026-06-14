package randomuser

import (
	"testing"
)

// These tests are offline: they exercise the URI driver's pure string functions.
// HTTP behaviour is covered in randomuser_test.go.

func TestDomainInfo(t *testing.T) {
	info := Domain{}.Info()
	if info.Scheme != "randomuser" {
		t.Errorf("Scheme = %q, want randomuser", info.Scheme)
	}
	if len(info.Hosts) == 0 || info.Hosts[0] != Host {
		t.Errorf("Hosts = %v, want [%s]", info.Hosts, Host)
	}
	if info.Identity.Binary != "randomuser" {
		t.Errorf("Identity.Binary = %q, want randomuser", info.Identity.Binary)
	}
}

func TestClassifyNumeric(t *testing.T) {
	typ, id, err := Domain{}.Classify("10")
	if err != nil {
		t.Fatalf("Classify(\"10\") error: %v", err)
	}
	if typ != "count" {
		t.Errorf("type = %q, want count", typ)
	}
	if id != "10" {
		t.Errorf("id = %q, want 10", id)
	}
}

func TestClassifyNationality(t *testing.T) {
	cases := []struct{ in, typ, id string }{
		{"us", "nationality", "us"},
		{"gb,au", "nationality", "gb,au"},
		{"US", "nationality", "US"},
	}
	for _, tc := range cases {
		typ, id, err := Domain{}.Classify(tc.in)
		if err != nil || typ != tc.typ || id != tc.id {
			t.Errorf("Classify(%q) = (%q, %q, %v), want (%q, %q, nil)",
				tc.in, typ, id, err, tc.typ, tc.id)
		}
	}
}

func TestClassifyEmpty(t *testing.T) {
	_, _, err := Domain{}.Classify("")
	if err == nil {
		t.Error("expected error for empty input, got nil")
	}
}

func TestLocateCount(t *testing.T) {
	got, err := Domain{}.Locate("count", "5")
	if err != nil {
		t.Fatalf("Locate(count): %v", err)
	}
	if got != "https://randomuser.me" {
		t.Errorf("Locate(count) = %q, want https://randomuser.me", got)
	}
}

func TestLocateNationality(t *testing.T) {
	got, err := Domain{}.Locate("nationality", "us")
	if err != nil {
		t.Fatalf("Locate(nationality): %v", err)
	}
	if got != "https://randomuser.me" {
		t.Errorf("Locate(nationality) = %q, want https://randomuser.me", got)
	}
}

func TestLocateUnknownType(t *testing.T) {
	_, err := Domain{}.Locate("unknown", "foo")
	if err == nil {
		t.Error("expected error for unknown type, got nil")
	}
}

func TestIsNumeric(t *testing.T) {
	cases := []struct {
		s    string
		want bool
	}{
		{"123", true},
		{"0", true},
		{"abc", false},
		{"12a", false},
		{"", false},
		{"us", false},
	}
	for _, tc := range cases {
		got := isNumeric(tc.s)
		if got != tc.want {
			t.Errorf("isNumeric(%q) = %v, want %v", tc.s, got, tc.want)
		}
	}
}
