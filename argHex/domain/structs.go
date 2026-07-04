package domain

import "time"

type Course struct {
	Institution Institution `json:"institution" bson:"institution,omitempty"`
	Name        string      `json:"name" bson:"name,omitempty"`
}

type Experience struct {
	Work       Institution `json:"work" bson:"work,omitempty"`
	Title      string      `json:"title" bson:"title,omitempty"`
	StartDate  Date        `json:"startDate" bson:"startDate,omitempty"`
	EndDate    Date        `json:"endDate" bson:"endDate,omitempty"`
	Highlights Highlights  `json:"highlights" bson:"highlights,omitempty"`
}

type CollegeExperience struct {
	School        Institution `json:"school" bson:"school,omitempty"`
	Degree        string      `json:"degree" bson:"degree,omitempty"`
	Major         string      `json:"major" bson:"major,omitempty"`
	CumulativeGPA float64     `json:"cumulativeGPA" bson:"cumulativeGPA,omitempty"`
	Highlights    Highlights  `json:"highlights" bson:"highlights,omitempty"`
	MajorGPA      float64     `json:"majorGPA" bson:"majorGPA,omitempty"`
	StartDate     Date        `json:"startDate" bson:"startDate,omitempty"`
	EndDate       Date        `json:"endDate" bson:"endDate,omitempty"`
}

type Institution struct {
	Name string `json:"name" bson:"name,omitempty"`
	// City string `json:"institutionCity" bson:"institutionCity,omitempty"`
	// State string `json:"institutionState" bson:"institutionState,omitempty"`
}

type Date struct {
	Year  int        `json:"year" bson:"year,omitempty"`
	Month time.Month `json:"month" bson:"month,omitempty"`
}

type Snippet struct {
	Name string `json:"name" bson:"name,omitempty"`
	Code string `json:"code" bson:"code,omitempty"`
}

type Link struct {
	Type string `json:"type" bson:"type,omitempty"`
	Text string `json:"text" bson:"text,omitempty"`
	URL  string `json:"url" bson:"url,omitempty"`
}

type Picture struct {
	Original string `json:"original" bson:"original,omitempty"`
	Thumb    string `json:"thumb" bson:"thumb,omitempty"`
	Text     string `json:"text" bson:"text,omitempty"`
}

type SimpleImage struct {
	Source string `json:"src" bson:"source,omitempty"`
	Alt    string `json:"alt" bson:"alt,omitempty"`
}

type Feature struct {
	Feature       string    `json:"feature" bson:"feature,omitempty"`
	IsComplete    bool      `json:"isComplete" bson:"isComplete,omitempty"`
	CompletedDate time.Time `json:"completedDate" bson:"completedDate,omitempty"`
}

type Highlight struct {
	Description string `json:"description" bson:"description,omitempty"`
	Visible     bool   `json:"visible" bson:"visible,omitempty"`
}

type SkillSection struct {
	Name   string `json:"name" bson:"name,omitempty"`
	Skills Skills `json:"skills" bson:"skills,omitempty"`
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

type Screenshot struct {
	Order int         `json:"order" bson:"order,omitempty"`
	Image SimpleImage `json:"image" bson:"image,omitempty"`
}

type HeroImage struct {
	Image SimpleImage `json:"image" bson:"image,omitempty"`
}

type SkillSections []SkillSection
type Highlights []Highlight
type Courses []Course
type Snippets []Snippet
type Links []Link
type Pictures []Picture
type Experiences []Experience
type Features []Feature
type Education []CollegeExperience
type TechInterests []TechInterest
type Contacts []Contact
type HeroImages []HeroImage
