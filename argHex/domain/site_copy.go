package domain

// SiteCopy is the "signal flags" singleton — the little lines of copy that fly
// over every page. There is exactly one document; it is upserted, never listed.
type SiteCopy struct {
	Id             string          `json:"id" bson:"_id,omitempty"`
	QuipHello      string          `json:"quipHello" bson:"quipHello,omitempty"`
	QuipProjects   string          `json:"quipProjects" bson:"quipProjects,omitempty"`
	QuipHobbies    string          `json:"quipHobbies" bson:"quipHobbies,omitempty"`
	QuipNotes      string          `json:"quipNotes" bson:"quipNotes,omitempty"`
	Quip404        string          `json:"quip404" bson:"quip404,omitempty"`
	HeroKicker     string          `json:"heroKicker" bson:"heroKicker,omitempty"`
	HeroHeadline   string          `json:"heroHeadline" bson:"heroHeadline,omitempty"`
	HeroBody       string          `json:"heroBody" bson:"heroBody,omitempty"`
	Dict           string          `json:"dict" bson:"dict,omitempty"`
	Eggs           *Eggs           `json:"eggs,omitempty" bson:"eggs,omitempty"`
	CatPages       map[string]bool `json:"catPages,omitempty" bson:"catPages,omitempty"`
	CatSpots       map[string]bool `json:"catSpots,omitempty" bson:"catSpots,omitempty"`
	BottleProverbs []string        `json:"bottleProverbs" bson:"bottleProverbs,omitempty"`
	Lighthouses    []Lighthouse    `json:"lighthouses" bson:"lighthouses,omitempty"`
	UpdatedAt      string          `json:"updatedAt" bson:"updatedAt,omitempty"`
}

// Eggs are the master switches for the easter eggs. The struct is a pointer on
// SiteCopy so legacy docs round-trip without it; consumers treat a missing
// block as everything-on. The bools deliberately skip omitempty — dropping a
// false would resurrect a switched-off egg on the next Save (see Hobby.Active).
type Eggs struct {
	Bottle bool `json:"bottle" bson:"bottle"`
	Cat    bool `json:"cat" bson:"cat"`
	Lights bool `json:"lights" bson:"lights"`
}

// CatPages and CatSpots are where the harbor cat is allowed to roam, keyed by
// page id (hello, projects, ...) and by `<page>.<spot>` spot id respectively.
// Maps rather than a struct because the keys are open-ended and live in the
// site/admin catalog, not here. The absent-means-on contract lives in the
// consumers; the API stores whatever it is handed. Bools inside a map serialize
// their false fine, so the omitempty-on-bool trap that bites Eggs can't reach
// them — the field-level omitempty only drops an empty or nil map.

// Lighthouse is one entry in the light list: a real light, its coordinates
// (the 404 wreck's "last position"), and the line it introduces itself with.
type Lighthouse struct {
	Name string `json:"name" bson:"name"`
	Pos  string `json:"pos" bson:"pos"`
	Line string `json:"line" bson:"line"`
}
