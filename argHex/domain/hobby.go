package domain

type Hobbies []Hobby

// Hobby is a "graveyard" entry: currently-learning (active) or resting. Order
// is a manual sort key so the keeper can arrange the headstones by hand.
type Hobby struct {
	Id        string `json:"id" bson:"_id,omitempty"`
	Name      string `json:"name" bson:"name,omitempty"`
	Dates     string `json:"dates" bson:"dates,omitempty"` // freeform display string
	Active    bool   `json:"active" bson:"active"`
	Epitaph   string `json:"epitaph" bson:"epitaph,omitempty"`
	Eulogy    string `json:"eulogy" bson:"eulogy,omitempty"`
	Order     int    `json:"order" bson:"order"`
	CreatedAt string `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt" bson:"updatedAt,omitempty"`
}
