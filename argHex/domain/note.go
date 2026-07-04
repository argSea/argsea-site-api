package domain

type Notes []Note

// Note is a writing-desk entry. Body is long-form rich text stored as a
// sanitized HTML string (banked decision); Image is a nullable media name.
type Note struct {
	Id          string  `json:"id" bson:"_id,omitempty"`
	Title       string  `json:"title" bson:"title,omitempty"`
	Teaser      string  `json:"teaser" bson:"teaser,omitempty"`
	Body        string  `json:"body" bson:"body,omitempty"` // sanitized HTML
	Date        string  `json:"date" bson:"date,omitempty"` // freeform display string
	Image       *string `json:"image" bson:"image,omitempty"`
	Status      string  `json:"status" bson:"status,omitempty"`
	PublishedAt string  `json:"publishedAt" bson:"publishedAt"` // no omitempty: unpublish must clear it
	CreatedAt   string  `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt   string  `json:"updatedAt" bson:"updatedAt,omitempty"`
}
