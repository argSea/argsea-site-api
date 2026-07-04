package domain

// Shared value objects still referenced by the user domain. The legacy
// portfolio/resume structs (Course, Experience, Snippet, Link, Feature, …) were
// removed with the old content model.

type SimpleImage struct {
	Source string `json:"src" bson:"source,omitempty"`
	Alt    string `json:"alt" bson:"alt,omitempty"`
}

type TechInterest struct {
	Name          string      `json:"name" bson:"name,omitempty"`
	Icon          SimpleImage `json:"icon" bson:"icon,omitempty"`
	InterestLevel int         `json:"interestLevel" bson:"interestLevel,omitempty"`
}

type Contact struct {
	Name string      `json:"name" bson:"name,omitempty"`
	Link string      `json:"link" bson:"link,omitempty"`
	Icon SimpleImage `json:"icon" bson:"icon,omitempty"`
}

type HeroImage struct {
	Image SimpleImage `json:"image" bson:"image,omitempty"`
}

type TechInterests []TechInterest
type Contacts []Contact
type HeroImages []HeroImage
