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
// LastLog carry the dates. Tags survive from the old shape: the home
// currently-learning card renders them, and the migration never touches them.
// NoteIds ties journal entries to this ship by stable Note id. Order is a
// manual sort key so the keeper arranges the log by hand. Gauge is the ship's
// self-assessed enthusiasm 0-100, clamped on write; a pointer so an older
// hobby with no opinion stays unset rather than reading as zero. Plate and Cap
// are the chart dressing, meaningful whether or not the ship is charted, and
// Images is the gallery the plate indexes into: a project's gallery shape
// exactly, first print leading by convention and capped at six.
type Hobby struct {
	Id        string   `json:"id" bson:"_id,omitempty"`
	Name      string   `json:"name" bson:"name,omitempty"`
	Service   string   `json:"service" bson:"service,omitempty"`
	State     string   `json:"state" bson:"state"`
	Coord     *Coord   `json:"coord" bson:"coord"`
	From      *Coord   `json:"from" bson:"from"`
	Plate     int      `json:"plate" bson:"plate"`             // no omitempty: 0 is the undressed plate and clearing one must survive a replace write
	Cap       string   `json:"cap" bson:"cap"`                 // no omitempty: empty is no caption and clearing one must survive a replace write
	Images    []string `json:"images" bson:"images,omitempty"` // gallery media names, capped at 6, first entry is the entry photo
	Seasons   string   `json:"seasons" bson:"seasons"`
	Bearing   string   `json:"bearing" bson:"bearing,omitempty"`
	LastLog   string   `json:"lastLog" bson:"lastLog,omitempty"`
	OffCourse string   `json:"offCourse" bson:"offCourse,omitempty"`
	Floats    string   `json:"floats" bson:"floats,omitempty"`
	Odds      string   `json:"odds" bson:"odds,omitempty"`
	Tags      []string `json:"tags" bson:"tags,omitempty"`
	NoteIds   []string `json:"noteIds" bson:"noteIds,omitempty"` // tied journal entries, by stable Note id
	Order     int      `json:"order" bson:"order"`
	Gauge     *int     `json:"gauge,omitempty" bson:"gauge,omitempty"` // nullable: self-assessed enthusiasm 0-100, absent means never rated
	CreatedAt string   `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt string   `json:"updatedAt" bson:"updatedAt,omitempty"`
}

// Coord is a point on the wandering chart. Lat/Lon are plain floats over the
// keeper's fictional waters. One rule guards every coord, whoever owns it: it
// has to be a real point on earth (see CoordOnEarth). Which frame a berth
// renders in is the site's business.
type Coord struct {
	Lat float64 `json:"lat" bson:"lat"`
	Lon float64 `json:"lon" bson:"lon"`
}

// The earth's own bounds. Nothing to do with any chart window: these are the
// only limits a latitude and a longitude have as numbers.
const (
	earthLatMin = -90.0
	earthLatMax = 90.0
	earthLonMin = -180.0
	earthLonMax = 180.0
)

// CoordOnEarth reports whether a coord is a real point on the globe, the shape
// check every chartable's bearing gets on write. It deliberately does not
// look at any chart window: a chart window is presentation, the site decides
// which frame a berth renders in, and validating one here would silently move a
// berth that sits outside it. A nil coord is uncharted, which is valid.
func CoordOnEarth(c *Coord) bool {
	if nil == c {
		return true
	}

	if earthLatMin > c.Lat || earthLatMax < c.Lat {
		return false
	}

	return earthLonMin <= c.Lon && earthLonMax >= c.Lon
}

// ValidHobbyState reports whether state is one the log allows. Empty is not a
// state: every ship stands somewhere.
func ValidHobbyState(state string) bool {
	return hobbyStates[state]
}

// The gauge's clamp band: a self-assessed enthusiasm reads 0-100, nothing else.
const (
	gaugeMin = 0
	gaugeMax = 100
)

// ClampGauge snaps a hobby's enthusiasm into 0-100, the same way ClampPlate
// snaps a plate index up to zero. A nil gauge is unrated and stays
// nil: the sanctioned no-opinion state, distinct from a gauge of 0.
func ClampGauge(g *int) {
	if nil == g {
		return
	}

	if gaugeMin > *g {
		*g = gaugeMin
	}

	if gaugeMax < *g {
		*g = gaugeMax
	}
}

// The plate's floor. There is deliberately no ceiling: the site owns how many
// photo plates exist and resolves an unknown index to its fallback plate, so an
// upper clamp here would be the API guessing the site's plate count and going
// stale the moment the site gains a plate.
const plateMin = 0

// ClampPlate snaps a chart plate index up to zero, the way ClampGauge snaps an
// enthusiasm into its band. Shared by every chartable: the dressing follows the
// entity while the coord decides chart presence, so a plate is clamped whether
// or not the entity is charted.
func ClampPlate(p *int) {
	if plateMin > *p {
		*p = plateMin
	}
}
