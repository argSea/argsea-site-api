package domain_test

import (
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
)

func TestPortBucketSortsReferrers(t *testing.T) {
	cases := []struct {
		name string
		ref  string
		want string
	}{
		{"empty referrer is direct", "", domain.PortDirect},
		{"google is search", "https://www.google.com/search?q=argsea", domain.PortSearch},
		{"duckduckgo is search", "https://duckduckgo.com/", domain.PortSearch},
		{"kagi is search", "https://kagi.com/search", domain.PortSearch},
		{"mastodon is fediverse", "https://mastodon.social/@someone/123", domain.PortFediverse},
		{"hachyderm is fediverse", "https://hachyderm.io/@keeper", domain.PortFediverse},
		{"a plain site is other", "https://news.ycombinator.com/item?id=1", domain.PortOther},
		{"a bare word ref is other", "not-a-url", domain.PortOther},
	}

	for _, c := range cases {
		if got := domain.PortBucket(c.ref); c.want != got {
			t.Fatalf("%s: PortBucket(%q) = %q, want %q", c.name, c.ref, got, c.want)
		}
	}
}

func TestVisitorHashVariesByDayAndIp(t *testing.T) {
	const (
		salt = "pepper"
		ua   = "Mozilla/5.0"
	)

	base := domain.VisitorHash(salt, "2026-07-12", "203.0.113.7", ua)

	if base != domain.VisitorHash(salt, "2026-07-12", "203.0.113.7", ua) {
		t.Fatalf("same inputs must hash to the same visitor")
	}

	if base == domain.VisitorHash(salt, "2026-07-13", "203.0.113.7", ua) {
		t.Fatalf("a different day must change the visitor hash")
	}

	if base == domain.VisitorHash(salt, "2026-07-12", "198.51.100.9", ua) {
		t.Fatalf("a different ip must change the visitor hash")
	}

	if base == domain.VisitorHash("other-salt", "2026-07-12", "203.0.113.7", ua) {
		t.Fatalf("a different salt must change the visitor hash")
	}
}

func TestVisitorHashIsTruncatedHex(t *testing.T) {
	hash := domain.VisitorHash("pepper", "2026-07-12", "203.0.113.7", "Mozilla/5.0")

	if 16 != len(hash) {
		t.Fatalf("expected a 16-char truncated hash, got %d: %q", len(hash), hash)
	}

	for _, r := range hash {
		if !('0' <= r && r <= '9' || 'a' <= r && r <= 'f') {
			t.Fatalf("visitor hash is not lowercase hex: %q", hash)
		}
	}
}

func TestIsBotCatchesCrawlersAndEmptyAgents(t *testing.T) {
	bots := []string{
		"",
		"Googlebot/2.1 (+http://www.google.com/bot.html)",
		"Some Crawler",
		"Twitterbot preview",
		"curl/8.4.0",
		"python-requests wget",
		"HeadlessChrome/120",
	}

	for _, ua := range bots {
		if !domain.IsBot(ua) {
			t.Fatalf("expected %q to be dropped as a bot", ua)
		}
	}

	humans := []string{
		"Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/605.1.15 Safari/605.1.15",
		"Mozilla/5.0 (X11; Linux x86_64) Gecko/20100101 Firefox/128.0",
	}

	for _, ua := range humans {
		if domain.IsBot(ua) {
			t.Fatalf("expected %q to pass as a human", ua)
		}
	}
}

func TestValidKind(t *testing.T) {
	for _, kind := range []string{domain.SightingSail, domain.SightingFlip, domain.SightingRead} {
		if !domain.ValidKind(kind) {
			t.Fatalf("expected %q to be a valid kind", kind)
		}
	}

	for _, kind := range []string{"", "click", "SAIL", "view"} {
		if domain.ValidKind(kind) {
			t.Fatalf("expected %q to be rejected as a kind", kind)
		}
	}
}

func TestValidPath(t *testing.T) {
	good := []string{"/", "/projects/foo", "/journal/2026/a-note"}

	for _, path := range good {
		if !domain.ValidPath(path) {
			t.Fatalf("expected %q to be a valid path", path)
		}
	}

	bad := []string{"", "projects/foo", "/has a space", "/has\ttab", "/has\nnewline"}

	for _, path := range bad {
		if domain.ValidPath(path) {
			t.Fatalf("expected %q to be rejected as a path", path)
		}
	}
}
