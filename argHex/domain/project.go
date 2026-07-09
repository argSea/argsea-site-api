package domain

type Projects []Project

// Stamp is the postage decoration in a postcard's top-right corner. Its
// vocabulary is deliberately closed: ink lands in style attributes on the
// public site and bluemonday only guards rich text, so the service-layer enum
// gate is the XSS boundary. An absent stamp is valid; the site renders its
// default decoration.
type Stamp struct {
	Shape string `json:"shape" bson:"shape,omitempty"` // rect | circle
	Motif string `json:"motif" bson:"motif,omitempty"` // lighthouse | boat | sun | wave | moon | anchor | text
	Ink   string `json:"ink" bson:"ink,omitempty"`     // #f0d9a8 | #93a0e8 (exact lowercase strings)
	Cents string `json:"cents" bson:"cents,omitempty"` // denomination "N¢" (rect shape only)
	Text  string `json:"text" bson:"text,omitempty"`   // caption: text motif only (required then), ≤ 40 chars after trim
}

// WallPos pins a light to an exact spot on the public coast panorama. X is
// the percentage along the shore, Y the percentage of elevation within the
// band (both 0-100); Rotation is a legacy tilt from the postcard wall that
// the coast view ignores but round-trips. No omitempty on these three: 0 is
// meaningful (far left / sea level / no tilt), same reason order and featured
// deliberately omit it on Project.
type WallPos struct {
	X        float64 `json:"x" bson:"x"`
	Y        float64 `json:"y" bson:"y"`
	Rotation float64 `json:"rotation" bson:"rotation"`
}

// WallPlacement is one entry in a bulk arrangement request: the light to pin
// and where on the coast it lands.
type WallPlacement struct {
	Id       string  `json:"id"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Rotation float64 `json:"rotation"`
}

// Light is how a project burns on the public coast: its navigational
// characteristic. Kind and Color are closed vocabularies gated in the service
// layer for the same reason as Stamp: they select animation names and glow
// colors rendered into style attributes on the public site, so the enum gate
// is the injection boundary. Period is the seconds one full cycle takes for
// the blinking kinds; a fixed light has none. Extinguished is a freeform year:
// any non-empty value means the light is dark (an abandoned project) while
// staying on the list. No omitempty on Period/Extinguished: 0 and "" are
// meaningful (fixed / still burning) and clearing them must survive a replace
// write.
type Light struct {
	Kind         string `json:"kind" bson:"kind,omitempty"`   // fixed | flash | occult | iso
	Color        string `json:"color" bson:"color,omitempty"` // white | red | green
	Period       int    `json:"period" bson:"period"`
	Extinguished string `json:"extinguished" bson:"extinguished"`
}

// Project is a light on the keeper's coast: the portfolio entry rendered as a
// navigational light on the public site (formerly a "postcard from
// production"). Body is long-form rich text stored as a sanitized HTML string
// (banked decision). Images is the entry's photo gallery, first print leads;
// Image is the postcard-era single photo the site falls back to when Images
// is empty. PostcardTo/PostcardFrom/Postmarked/Stamp are dormant postcard-era
// fields: no longer written by the admin, preserved so old documents and
// revisions stay readable.
type Project struct {
	Id           string   `json:"id" bson:"_id,omitempty"`
	Title        string   `json:"title" bson:"title,omitempty"`
	Category     string   `json:"category" bson:"category,omitempty"` // backend | games | this website | tinkering
	Tags         []string `json:"tags" bson:"tags,omitempty"`
	ShortDesc    string   `json:"shortDesc" bson:"shortDesc,omitempty"` // "front of card"
	Body         string   `json:"body" bson:"body,omitempty"`           // sanitized HTML long-form
	Moral        string   `json:"moral" bson:"moral,omitempty"`
	PostcardTo   string   `json:"postcardTo" bson:"postcardTo,omitempty"`
	PostcardFrom string   `json:"postcardFrom" bson:"postcardFrom,omitempty"`
	Postmarked   string   `json:"postmarked" bson:"postmarked,omitempty"` // freeform display string
	Slug         string   `json:"slug" bson:"slug,omitempty"`
	Image        *string  `json:"image" bson:"image,omitempty"`           // nullable media name (legacy single photo)
	Stamp        *Stamp   `json:"stamp" bson:"stamp,omitempty"`           // nullable postage decoration (dormant)
	Light        *Light   `json:"light,omitempty" bson:"light,omitempty"` // nullable: nil burns as the default fixed white
	Images       []string `json:"images" bson:"images,omitempty"`         // gallery media names, first print leads
	FirstLit     string   `json:"firstLit" bson:"firstLit,omitempty"`     // freeform year shown in the register
	Status       string   `json:"status" bson:"status,omitempty"`
	Order        int      `json:"order" bson:"order"`                         // no omitempty: 0 is a real rack position
	Featured     bool     `json:"featured" bson:"featured"`                   // no omitempty: false must survive a replace write
	PublishedAt  string   `json:"publishedAt" bson:"publishedAt"`             // no omitempty: unpublish must clear it
	WallPos      *WallPos `json:"wallPos,omitempty" bson:"wallPos,omitempty"` // nullable: nil means not yet placed on the wall
	CreatedAt    string   `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt    string   `json:"updatedAt" bson:"updatedAt,omitempty"`
}
