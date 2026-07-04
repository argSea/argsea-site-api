package domain

type Resumes []Resume

type Resume struct {
	Id            string        `json:"resumeID" bson:"_id,omitempty"`
	UserID        string        `json:"userID" bson:"userID,omitempty"`
	About         string        `json:"about" bson:"about,omitempty"`
	Experiences   Experiences   `json:"experiences" bson:"experiences,omitempty"`
	Education     Education     `json:"education" bson:"education,omitempty"`
	ExtraCourses  Courses       `json:"extraCourses" bson:"extraCourses,omitempty"`
	SkillSections SkillSections `json:"skills" bson:"skills,omitempty"`
}
