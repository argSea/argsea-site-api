package domain

type MediaList []Media

// Media is a darkroom item. The metadata half lives in mongo; the file half
// sits on disk behind the webstore adapter. URL is the web-relative path the
// site serves the file from (web_path + filename); CreatedAt uses the same
// fixed-width RFC3339 stamp as the rest of the content model so newest-first
// is a plain string sort.
type Media struct {
	Id        string `json:"id" bson:"_id,omitempty"`
	Filename  string `json:"filename" bson:"filename,omitempty"`
	URL       string `json:"url" bson:"url,omitempty"`
	CreatedAt string `json:"createdAt" bson:"createdAt,omitempty"`
}
