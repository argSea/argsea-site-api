package domain

// Watch is the "current watch" singleton: the hand-written /now record on the
// homepage. There is exactly one document; it is upserted, never listed. An
// empty Letter is the cleared state (the site collapses the section), so
// clearing is just an authed write of an empty record; there is no delete.
type Watch struct {
	Id              string         `json:"id" bson:"_id,omitempty"`
	Letter          string         `json:"letter" bson:"letter,omitempty"`                   // hand-written; a blank line splits paragraphs
	Rotation        string         `json:"rotation" bson:"rotation,omitempty"`               // the not-doing line ("out of the rotation")
	Bearings        []WatchBearing `json:"bearings" bson:"bearings,omitempty"`               // the TL;DR strip; the service truncates past three
	PostcardMediaId string         `json:"postcardMediaId" bson:"postcardMediaId,omitempty"` // darkroom print id; "" means no postcard
	Quips           []string       `json:"quips" bson:"quips,omitempty"`                     // the cat's remarks on the watch panel
	KeptAt          string         `json:"keptAt" bson:"keptAt,omitempty"`                   // stamped server-side on save; a client-sent value is ignored
}

// WatchBearing is one line on the bearings strip: a verb and what it points at.
// Verb is freeform ("wrangling", "logging", "tinkering"). Kind is "none",
// "light", "hobby" or "note"; TargetId is empty when Kind is "none". The inner
// fields skip omitempty so a "none" bearing's empty TargetId survives the
// Replace-based Save intact.
type WatchBearing struct {
	Verb     string `json:"verb" bson:"verb"`
	Kind     string `json:"kind" bson:"kind"`
	TargetId string `json:"targetId" bson:"targetId"`
	Name     string `json:"name" bson:"name"`
}
