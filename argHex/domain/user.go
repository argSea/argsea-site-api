package domain

type Users []User

// Entity // domain
type User struct {
	//Model
	Id            string        `json:"id" bson:"_id,omitempty"`
	UserName      string        `json:"userName" bson:"userName,omitempty"`
	Password      Password      `json:"password" bson:"password,omitempty"`
	FirstName     string        `json:"firstName" bson:"firstName,omitempty"`
	LastName      string        `json:"lastName" bson:"lastName,omitempty"`
	Contacts      Contacts      `json:"contacts" bson:"contacts,omitempty"`
	Pictures      HeroImages    `json:"pictures" bson:"pictures,omitempty"`
	About         string        `json:"about" bson:"about,omitempty"`
	TechInterests TechInterests `json:"techInterests" bson:"techInterests,omitempty"`
	// keeper profile (operator ruling 2026-07-05: no separate keeper entity;
	// this data lives on the user and is served publicly through Profile).
	// No bson omitempty on the nine profile fields: user writes go through a
	// $set update, so the key must travel for a PUT that empties a field (say,
	// pronouns) to actually persist the clear; same rule as PublishedAt.
	Name     string `json:"name" bson:"name"`
	Pronouns string `json:"pronouns" bson:"pronouns"`
	Location string `json:"location" bson:"location"`
	Title    string `json:"title" bson:"title"`
	Bio      string `json:"bio" bson:"bio"`
	Email    string `json:"email" bson:"email"`
	Github   string `json:"github" bson:"github"`
	Linkedin string `json:"linkedin" bson:"linkedin"`
	Signoff  string `json:"signoff" bson:"signoff"`
	// Role is what login mints into the JWT. It is never accepted from a request
	// body; admin is granted only by a direct DB update on the user document.
	Role string `json:"role,omitempty" bson:"role,omitempty"`
}

// UserProfile is the public keeper subset served by GET /1/user/{id}/profile:
// the nine profile fields only, never username, password, or role.
type UserProfile struct {
	Name     string `json:"name"`
	Pronouns string `json:"pronouns"`
	Location string `json:"location"`
	Title    string `json:"title"`
	Bio      string `json:"bio"`
	Email    string `json:"email"`
	Github   string `json:"github"`
	Linkedin string `json:"linkedin"`
	Signoff  string `json:"signoff"`
}

// Profile projects the public keeper subset out of the user document.
func (u User) Profile() UserProfile {
	return UserProfile{
		Name:     u.Name,
		Pronouns: u.Pronouns,
		Location: u.Location,
		Title:    u.Title,
		Bio:      u.Bio,
		Email:    u.Email,
		Github:   u.Github,
		Linkedin: u.Linkedin,
		Signoff:  u.Signoff,
	}
}

type Password string

func (Password) MarshalJSON() ([]byte, error) {
	return []byte(`""`), nil
}
