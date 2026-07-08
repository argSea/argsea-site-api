package utility

import "github.com/microcosm-cc/bluemonday"

// htmlPolicy is the allowlist applied to rich-text body fields. UGCPolicy keeps
// the common formatting tags a user-generated post needs (headings, lists,
// links, code, emphasis) while stripping scripts, event handlers, and any other
// XSS vector. Built once; the policy is safe for concurrent use.
var htmlPolicy = bluemonday.UGCPolicy()

// SanitizeHTML returns a version of raw safe to store and later render verbatim.
// Rich-text bodies are a banked decision: they live in the database already
// sanitized, so the API is the trust boundary that guarantees it.
func SanitizeHTML(raw string) string {
	return htmlPolicy.Sanitize(raw)
}
