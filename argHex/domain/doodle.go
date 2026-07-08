package domain

type Doodles []Doodle

// Doodle is a marginalia sketch for the Keeper's Journal, drawn in the admin.
// Shapes are structured geometry, never markup (the same banked XSS decision
// as the figurehead designs); the renderers live in the admin and the site,
// the API only stores the JSON. Reuses the figurehead Shape type verbatim;
// Role/Origin simply go unused here.
type Doodle struct {
	Id        string  `json:"id" bson:"_id,omitempty"`
	Name      string  `json:"name" bson:"name,omitempty"`
	ViewBox   string  `json:"viewBox" bson:"viewBox,omitempty"`
	Shapes    []Shape `json:"shapes" bson:"shapes,omitempty"`
	CreatedAt string  `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt string  `json:"updatedAt" bson:"updatedAt,omitempty"`
}
