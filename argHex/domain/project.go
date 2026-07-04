package domain

type Projects []Project

// Stamp is the postage decoration in a postcard's top-right corner. Its
// vocabulary is deliberately closed: ink lands in style attributes on the
// public site and bluemonday only guards rich text, so the service-layer enum
// gate is the XSS boundary. An absent stamp is valid — the site renders its
// default decoration.
type Stamp struct {
	Shape string `json:"shape" bson:"shape,omitempty"` // rect | circle
	Motif string `json:"motif" bson:"motif,omitempty"` // lighthouse | boat | sun | wave | moon | anchor | text
	Ink   string `json:"ink" bson:"ink,omitempty"`     // #f0d9a8 | #93a0e8 (exact lowercase strings)
	Cents string `json:"cents" bson:"cents,omitempty"` // denomination shown on rect stamps, "N¢"
	Text  string `json:"text" bson:"text,omitempty"`   // caption for the text motif, ≤ 40 chars after trim
}

// Project is a "postcard from production" — the reshaped portfolio entry that
// replaces the legacy user-scoped project. Body is long-form rich text stored
// as a sanitized HTML string (banked decision); Image is a nullable reference
// to a media item by name.
type Project struct {
	Id           string   `json:"id" bson:"_id,omitempty"`
	Title        string   `json:"title" bson:"title,omitempty"`
	Category     string   `json:"category" bson:"category,omitempty"` // backend | games | this website | tinkering
	Tags         []string `json:"tags" bson:"tags,omitempty"`
	ShortDesc    string   `json:"shortDesc" bson:"shortDesc,omitempty"` // "front of card"
	Body         string   `json:"body" bson:"body,omitempty"`           // sanitized HTML long-form
	Moral        string   `json:"moral" bson:"moral,omitempty"`
	PostcardTo   string   `json:"postcardTo" bson:"postcardTo,omitempty"`
	PostcardFrom string   `json:"postcardFrom" bson:"postcardFrom,omitempty"`
	Postmarked   string   `json:"postmarked" bson:"postmarked,omitempty"` // freeform display string
	Slug         string   `json:"slug" bson:"slug,omitempty"`
	Image        *string  `json:"image" bson:"image,omitempty"` // nullable media name
	Stamp        *Stamp   `json:"stamp" bson:"stamp,omitempty"` // nullable postage decoration
	Status       string   `json:"status" bson:"status,omitempty"`
	PublishedAt  string   `json:"publishedAt" bson:"publishedAt"` // no omitempty: unpublish must clear it
	CreatedAt    string   `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt    string   `json:"updatedAt" bson:"updatedAt,omitempty"`
}
