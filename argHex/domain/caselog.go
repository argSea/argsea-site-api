package domain

type CaseLogs []CaseLog

// Block is one entry in a case study, a discriminated union keyed on its "kind"
// field. The API never interprets a block beyond that: blocks are stored and
// returned verbatim, so a map is the currency rather than a typed struct. That
// is deliberate, not lazy. Only the site parses a block's contents (including
// the keeper's inline marks inside text fields), and a typed struct with
// omitempty tags would silently drop a present-but-zero field like a list's
// ordered:false or an empty timeline link, breaking the wire contract three
// repos build against. The verbatim map cannot drop what it was handed.
type Block map[string]interface{}

type Blocks []Block

// CaseLog is a project's full case study as its own document: the long-form
// story that used to live in the dormant Project.caseStudy string, now a list
// of blocks. The header (title, subhead, facts, meta) lives in the blocks
// themselves. Title is a derived display title the admin keeps synced from the
// first title block, carried here so lists need not walk the blocks. At most
// one published log per project; the publish swap enforces it. ProjectId is
// required and references an existing project.
type CaseLog struct {
	Id          string `json:"id" bson:"_id,omitempty"`
	ProjectId   string `json:"projectId" bson:"projectId,omitempty"`
	Status      string `json:"status" bson:"status,omitempty"`
	Title       string `json:"title" bson:"title,omitempty"`
	Blocks      Blocks `json:"blocks" bson:"blocks,omitempty"`
	PublishedAt string `json:"publishedAt" bson:"publishedAt"` // no omitempty: unpublish must clear it
	CreatedAt   string `json:"createdAt" bson:"createdAt,omitempty"`
	UpdatedAt   string `json:"updatedAt" bson:"updatedAt,omitempty"`
}

// LegacyCaseStudy is the boot migration's read-only currency: one project's
// dormant caseStudy string plus the fields the header seed is built from. The
// live Project struct no longer carries caseStudy, so the projects collection
// is this shape's only source and nothing writes it back.
type LegacyCaseStudy struct {
	ProjectId string
	Title     string
	ShortDesc string
	FirstLit  string
	Tags      []string
	Facts     []ProjectFact
	CaseStudy string
}

type LegacyCaseStudies []LegacyCaseStudy
