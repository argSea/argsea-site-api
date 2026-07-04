package domain

type Skills []Skill

type Skill struct {
	Id          string `json:"id" bson:"_id,omitempty"`
	Name        string `json:"name" bson:"name,omitempty"`
	Description string `json:"description" bson:"description,omitempty"`
}
