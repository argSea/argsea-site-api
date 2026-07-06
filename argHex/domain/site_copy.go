package domain

// SiteCopy is the "signal flags" singleton — the little lines of copy that fly
// over every page. There is exactly one document; it is upserted, never listed.
type SiteCopy struct {
	Id             string       `json:"id" bson:"_id,omitempty"`
	QuipHello      string       `json:"quipHello" bson:"quipHello,omitempty"`
	QuipProjects   string       `json:"quipProjects" bson:"quipProjects,omitempty"`
	QuipHobbies    string       `json:"quipHobbies" bson:"quipHobbies,omitempty"`
	QuipNotes      string       `json:"quipNotes" bson:"quipNotes,omitempty"`
	Quip404        string       `json:"quip404" bson:"quip404,omitempty"`
	HeroKicker     string       `json:"heroKicker" bson:"heroKicker,omitempty"`
	HeroHeadline   string       `json:"heroHeadline" bson:"heroHeadline,omitempty"`
	HeroBody       string       `json:"heroBody" bson:"heroBody,omitempty"`
	Dict           string       `json:"dict" bson:"dict,omitempty"`
	Eggs           *Eggs        `json:"eggs,omitempty" bson:"eggs,omitempty"`
	CatLocs        *CatLocs     `json:"catLocs,omitempty" bson:"catLocs,omitempty"`
	BottleProverbs []string     `json:"bottleProverbs" bson:"bottleProverbs,omitempty"`
	Lighthouses    []Lighthouse `json:"lighthouses" bson:"lighthouses,omitempty"`
	UpdatedAt      string       `json:"updatedAt" bson:"updatedAt,omitempty"`
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

// CatLocs is where the harbor cat is allowed to roam. Same absent-means-on
// contract as Eggs, so the bools skip omitempty too.
type CatLocs struct {
	Postcards bool `json:"postcards" bson:"postcards"`
	Notes     bool `json:"notes" bson:"notes"`
	P404      bool `json:"p404" bson:"p404"`
}

// Lighthouse is one entry in the light list: a real light, its coordinates
// (the 404 wreck's "last position"), and the line it introduces itself with.
type Lighthouse struct {
	Name string `json:"name" bson:"name"`
	Pos  string `json:"pos" bson:"pos"`
	Line string `json:"line" bson:"line"`
}
