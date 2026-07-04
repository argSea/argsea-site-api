package domain

type Projects []Project

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
	Status       string   `json:"status" bson:"status,omitempty"`
	PublishedAt  string   `json:"publishedAt" bson:"publishedAt"` // no omitempty: unpublish must clear it
	CreatedAt    string   `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt    string   `json:"updatedAt" bson:"updatedAt,omitempty"`
}
