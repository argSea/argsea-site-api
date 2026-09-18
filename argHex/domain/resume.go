package domain

import "strings"

type Resumes []Resume

// Resume is one cut of the keeper's papers: a stored PDF plus the title and
// notes he writes for his own use, so a Senior Software Engineer cut sits on
// the shelf beside a Systems Architect cut and he can choose between them.
// Published is the whole lifecycle and exactly one cut carries it; draft is the
// absence of publication, not a state of its own. Filename and URL are stamped
// by the service from the generated name the upload landed under, never sent by
// a client, because the payload is immutable once stored. CreatedAt/UpdatedAt
// use the same fixed-width RFC3339 stamp as the rest of the content model so
// newest-first is a plain string sort.
type Resume struct {
	Id        string `json:"id" bson:"_id,omitempty"`
	Title     string `json:"title" bson:"title,omitempty"`
	Notes     string `json:"notes" bson:"notes,omitempty"`
	Filename  string `json:"filename" bson:"filename,omitempty"`
	URL       string `json:"url" bson:"url,omitempty"`
	Published bool   `json:"published" bson:"published"` // no omitempty: false is the unpublished state and clearing one must survive a replace write
	CreatedAt string `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt" bson:"updatedAt,omitempty"`
}

// ValidResumeTitle reports whether a cut carries the title the keeper picks it
// by. Every stored resume is the same kind of PDF under a generated name, so an
// untitled one is indistinguishable from the rest of the shelf.
func ValidResumeTitle(title string) bool {
	return "" != strings.TrimSpace(title)
}
