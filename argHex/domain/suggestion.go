package domain

type Suggestions []Suggestion

// Suggestion is one chip in the hobby "next: ???" pool. Order keeps the pool
// in the sequence the keeper added (or later rearranged) them.
type Suggestion struct {
	Id    string `json:"id" bson:"_id,omitempty"`
	Value string `json:"value" bson:"value,omitempty"`
	Order int    `json:"order" bson:"order"`
}
