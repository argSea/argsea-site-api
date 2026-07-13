package domain

type Hobbies []Hobby

// hobbyStates is the closed vocabulary of ways a ship stands in the log: moored
// at its berth, made port elsewhere, adrift, marooned, or its bearing smudged to
// an inkspill. It gates state the same way the light's kind gates a project.
var hobbyStates = map[string]bool{
	StateMoored:   true,
	StatePort:     true,
	StateAdrift:   true,
	StateMarooned: true,
	StateInkspill: true,
}

// the five states a ship in the log can stand in. A hobby always carries one;
// there is no empty state, unlike an absent stamp or light.
const (
	StateMoored   = "moored"
	StatePort     = "port"
	StateAdrift   = "adrift"
	StateMarooned = "marooned"
	StateInkspill = "inkspill"
)

// Hobby is one ship in the ship's log: a pursuit at its last known bearing on
// the wandering chart. State is the closed vocabulary above, validated on write.
// Coord is where the ship sits on the chart and From is the wake it trailed in
// on; both are pointers so an uncharted ship serializes coord/from as JSON null
// rather than a phantom origin at 0,0. Seasons is a free string ("5", "¼", "").
// Bearing, OffCourse, Floats, and Odds are the log's freeform prose; Service and
// LastLog carry the dates. Order is a manual sort key so the keeper arranges the
// log by hand.
type Hobby struct {
	Id        string `json:"id" bson:"_id,omitempty"`
	Name      string `json:"name" bson:"name,omitempty"`
	Service   string `json:"service" bson:"service,omitempty"`
	State     string `json:"state" bson:"state"`
	Coord     *Coord `json:"coord" bson:"coord"`
	From      *Coord `json:"from" bson:"from"`
	Seasons   string `json:"seasons" bson:"seasons"`
	Bearing   string `json:"bearing" bson:"bearing,omitempty"`
	LastLog   string `json:"lastLog" bson:"lastLog,omitempty"`
	OffCourse string `json:"offCourse" bson:"offCourse,omitempty"`
	Floats    string `json:"floats" bson:"floats,omitempty"`
	Odds      string `json:"odds" bson:"odds,omitempty"`
	Order     int    `json:"order" bson:"order"`
	CreatedAt string `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt string `json:"updatedAt" bson:"updatedAt,omitempty"`
}

// Coord is a point on the wandering chart. Lat/Lon are plain floats: the keeper
// charts fictional waters, so there is no range to validate beyond being finite.
type Coord struct {
	Lat float64 `json:"lat" bson:"lat"`
	Lon float64 `json:"lon" bson:"lon"`
}

// ValidHobbyState reports whether state is one the log allows. Empty is not a
// state: every ship stands somewhere.
func ValidHobbyState(state string) bool {
	return hobbyStates[state]
}
