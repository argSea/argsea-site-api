package domain

type Hobbies []Hobby

// Hobby is a "graveyard" entry: currently-learning (active) or resting. Order
// is a manual sort key so the keeper can arrange the headstones by hand.
// Service/Char/Disposition/Log/LastLog/Found/Cause/Return are the register's
// freeform prose fields. Marker is the headstone shape, closed vocabulary
// like the light's kind; Wear is its weathering, a fraction like the light's
// period is a bounded number. Dates/Epitaph/Eulogy are dormant postcard-era
// fields: no longer written by the admin, preserved so old documents and
// revisions stay readable (same precedent as Project's dormant fields).
type Hobby struct {
	Id          string   `json:"id" bson:"_id,omitempty"`
	Name        string   `json:"name" bson:"name,omitempty"`
	Dates       string   `json:"dates" bson:"dates,omitempty"` // dormant
	Active      bool     `json:"active" bson:"active"`
	Epitaph     string   `json:"epitaph" bson:"epitaph,omitempty"` // dormant
	Eulogy      string   `json:"eulogy" bson:"eulogy,omitempty"`   // dormant
	Tags        []string `json:"tags" bson:"tags,omitempty"`
	Order       int      `json:"order" bson:"order"`
	Service     string   `json:"service" bson:"service,omitempty"`
	Char        string   `json:"char" bson:"char,omitempty"`
	Disposition string   `json:"disposition" bson:"disposition,omitempty"`
	Log         string   `json:"log" bson:"log,omitempty"`
	LastLog     string   `json:"lastLog" bson:"lastLog,omitempty"`
	Found       string   `json:"found" bson:"found,omitempty"`
	Cause       string   `json:"cause" bson:"cause,omitempty"`
	Return      string   `json:"return" bson:"return,omitempty"`
	Marker      string   `json:"marker" bson:"marker,omitempty"` // stone | sticks | driftwood | cairn | buoy | lamp
	Wear        float64  `json:"wear" bson:"wear"`               // no omitempty: 0 is real weathering; 0.0-1.0
	CreatedAt   string   `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt   string   `json:"updatedAt" bson:"updatedAt,omitempty"`
}
