package domain

type CatDesigns []CatDesign

// The two stances the harbor cat holds around the site. A design dresses
// exactly one of them.
const (
	PosePerched = "perched"
	PoseLying   = "lying"
)

// CatDesign is one figurehead outfit for the harbor cat, drawn in the admin's
// Figurehead Shop. Shapes are structured geometry, never markup (banked XSS
// decision) — the renderers live in the admin and the site, the API only
// stores the JSON. Exactly one design per pose is published at a time;
// publishing another supersedes it.
type CatDesign struct {
	Id        string  `json:"id" bson:"_id,omitempty"`
	Pose      string  `json:"pose" bson:"pose,omitempty"` // perched | lying
	Label     string  `json:"label" bson:"label,omitempty"`
	ViewBox   string  `json:"viewBox" bson:"viewBox,omitempty"`
	Shapes    []Shape `json:"shapes" bson:"shapes,omitempty"`
	Published bool    `json:"published" bson:"published"` // no omitempty: an explicit false must survive a replace write
	Seed      bool    `json:"seed" bson:"seed"`           // no omitempty: same trap — the v1 flag must never resurrect or vanish
	CreatedAt string  `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt string  `json:"updatedAt" bson:"updatedAt,omitempty"`
}

// Shape is one SVG primitive of a design. Only the geometry fields matching
// Type carry meaning (d for a path, cx/cy/rx/ry for an ellipse, x/y/w/h for a
// rect, x1/y1/x2/y2 for a line); everything optional is omitempty, and an
// absent field means the SVG attribute default — renderers write only the
// fields present. Role and Origin drive the site's canonical animations (tail
// sway, blink); the API stores them opaquely.
type Shape struct {
	Id          string    `json:"id,omitempty" bson:"id,omitempty"`
	Type        string    `json:"type" bson:"type,omitempty"` // path | ellipse | rect | line
	D           string    `json:"d,omitempty" bson:"d,omitempty"`
	Cx          float64   `json:"cx,omitempty" bson:"cx,omitempty"`
	Cy          float64   `json:"cy,omitempty" bson:"cy,omitempty"`
	Rx          float64   `json:"rx,omitempty" bson:"rx,omitempty"`
	Ry          float64   `json:"ry,omitempty" bson:"ry,omitempty"`
	X           float64   `json:"x,omitempty" bson:"x,omitempty"`
	Y           float64   `json:"y,omitempty" bson:"y,omitempty"`
	W           float64   `json:"w,omitempty" bson:"w,omitempty"`
	H           float64   `json:"h,omitempty" bson:"h,omitempty"`
	X1          float64   `json:"x1,omitempty" bson:"x1,omitempty"`
	Y1          float64   `json:"y1,omitempty" bson:"y1,omitempty"`
	X2          float64   `json:"x2,omitempty" bson:"x2,omitempty"`
	Y2          float64   `json:"y2,omitempty" bson:"y2,omitempty"`
	Fill        string    `json:"fill,omitempty" bson:"fill,omitempty"`
	Stroke      string    `json:"stroke,omitempty" bson:"stroke,omitempty"`
	StrokeWidth float64   `json:"strokeWidth,omitempty" bson:"strokeWidth,omitempty"`
	Opacity     float64   `json:"opacity,omitempty" bson:"opacity,omitempty"`
	Linecap     string    `json:"linecap,omitempty" bson:"linecap,omitempty"`
	Linejoin    string    `json:"linejoin,omitempty" bson:"linejoin,omitempty"`
	Role        string    `json:"role,omitempty" bson:"role,omitempty"`     // tail | eyes | body — stored opaquely
	Origin      []float64 `json:"origin,omitempty" bson:"origin,omitempty"` // [x, y] transform origin for the role's animation
}
