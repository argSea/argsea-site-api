package domain

import "time"

type Media struct {
	ID        string    `json:"id" bson:"_id,omitempty"`
	Title     string    `json:"title" bson:"title,omitempty"`
	Filename  string    `json:"filename" bson:"filename,omitempty"`
	URL       string    `json:"url" bson:"url,omitempty"`
	CreatedAt time.Time `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt time.Time `json:"updatedAt" bson:"updatedAt,omitempty"`
}
