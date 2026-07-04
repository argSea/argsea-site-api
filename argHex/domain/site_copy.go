package domain

// SiteCopy is the "signal flags" singleton — the little lines of copy that fly
// over every page. There is exactly one document; it is upserted, never listed.
type SiteCopy struct {
	Id           string `json:"id" bson:"_id,omitempty"`
	QuipHello    string `json:"quipHello" bson:"quipHello,omitempty"`
	QuipProjects string `json:"quipProjects" bson:"quipProjects,omitempty"`
	QuipHobbies  string `json:"quipHobbies" bson:"quipHobbies,omitempty"`
	QuipNotes    string `json:"quipNotes" bson:"quipNotes,omitempty"`
	Quip404      string `json:"quip404" bson:"quip404,omitempty"`
	HeroKicker   string `json:"heroKicker" bson:"heroKicker,omitempty"`
	HeroHeadline string `json:"heroHeadline" bson:"heroHeadline,omitempty"`
	HeroBody     string `json:"heroBody" bson:"heroBody,omitempty"`
	Dict         string `json:"dict" bson:"dict,omitempty"`
	UpdatedAt    string `json:"updatedAt" bson:"updatedAt,omitempty"`
}
