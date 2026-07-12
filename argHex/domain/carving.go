package domain

type Carvings []Carving

// The seven spots a carving can bolt onto: every hand-carved svg on the site
// minus the doodles (marginalia sketches never take a carving) and the
// catalog-only entries that live admin-side as display notes (the postage
// stamp, the wreck, the harbor cat itself). The set is frozen; a new spot
// needs its own change, not a config toggle.
const (
	SpotLighthouseLogo = "lighthouse-logo"
	SpotBoat           = "boat"
	SpotBottle         = "bottle"
	SpotTowerStub      = "tower-stub"
	SpotPaw            = "paw"
	SpotWaveLine       = "wave-line"
	SpotBoatWake       = "boat-wake"
)

// CarvingSpots is the closed spot vocabulary: bolting anywhere outside it
// would target an element the site never renders.
var CarvingSpots = map[string]bool{
	SpotLighthouseLogo: true,
	SpotBoat:           true,
	SpotBottle:         true,
	SpotTowerStub:      true,
	SpotPaw:            true,
	SpotWaveLine:       true,
	SpotBoatWake:       true,
}

// Carving is one raw-svg block cut at the bench, bolted onto zero or more
// spots. Builtin marks the seven shipped v1 seeds: their Name and Svg are
// frozen so every spot always has a v1 to bolt back to, but BoltedTo stays
// mutable even on a builtin, since re-bolting a spot to its own v1 must
// always be possible. Exactly one carving holds a given spot at a time;
// bolting it elsewhere strips it from whoever held it.
type Carving struct {
	Id        string   `json:"id" bson:"_id,omitempty"`
	Name      string   `json:"name" bson:"name,omitempty"`
	Svg       string   `json:"svg" bson:"svg,omitempty"`           // raw SVG markup
	Builtin   bool     `json:"builtin" bson:"builtin"`             // v1 seed; name+svg frozen
	BoltedTo  []string `json:"boltedTo" bson:"boltedTo,omitempty"` // spot ids
	CreatedAt string   `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt string   `json:"updatedAt" bson:"updatedAt,omitempty"`
}
