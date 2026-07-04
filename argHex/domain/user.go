package domain

type Users []User

//Entity // domain
type User struct {
	//Model
	Id            string        `json:"id" bson:"_id,omitempty"`
	UserName      string        `json:"userName" bson:"userName,omitempty"`
	Password      Password      `json:"password" bson:"password,omitempty"`
	FirstName     string        `json:"firstName" bson:"firstName,omitempty"`
	LastName      string        `json:"lastName" bson:"lastName,omitempty"`
	Email         string        `json:"email" bson:"email,omitempty"`
	Contacts      Contacts      `json:"contacts" bson:"contacts,omitempty"`
	Title         string        `json:"title" bson:"title,omitempty"`
	Pictures      HeroImages    `json:"pictures" bson:"pictures,omitempty"`
	About         string        `json:"about" bson:"about,omitempty"`
	TechInterests TechInterests `json:"techInterests" bson:"techInterests,omitempty"`
}

type Password string

func (Password) MarshalJSON() ([]byte, error) {
	return []byte(`""`), nil
}
