package domain

type Notes []Note

// Note is a writing-desk entry. Body is long-form rich text stored as a
// sanitized HTML string (banked decision); DoodleId is a nullable reference to
// a Doodle vector resource. Coord is where the entry sits on the wandering
// chart, with Hobby.Coord's semantics: a pointer so an uncharted entry
// serializes coord as JSON null rather than a phantom origin at 0,0. Plate and
// Cap are the chart dressing, meaningful whether or not the entry is charted.
type Note struct {
	Id            string  `json:"id" bson:"_id,omitempty"`
	Title         string  `json:"title" bson:"title,omitempty"`
	Teaser        string  `json:"teaser" bson:"teaser,omitempty"`
	Body          string  `json:"body" bson:"body,omitempty"` // sanitized HTML
	Date          string  `json:"date" bson:"date,omitempty"` // freeform display string
	Conditions    string  `json:"conditions" bson:"conditions,omitempty"`
	DoodleCaption string  `json:"doodleCaption" bson:"doodleCaption,omitempty"`
	DoodleId      *string `json:"doodleId" bson:"doodleId,omitempty"`
	Coord         *Coord  `json:"coord" bson:"coord"` // nullable: null means uncharted, in the log and off the chart
	Plate         int     `json:"plate" bson:"plate"` // no omitempty: 0 is the undressed plate and clearing one must survive a replace write
	Cap           string  `json:"cap" bson:"cap"`     // no omitempty: empty is no caption and clearing one must survive a replace write
	Status        string  `json:"status" bson:"status,omitempty"`
	PublishedAt   string  `json:"publishedAt" bson:"publishedAt"` // no omitempty: unpublish must clear it
	CreatedAt     string  `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt     string  `json:"updatedAt" bson:"updatedAt,omitempty"`
}
