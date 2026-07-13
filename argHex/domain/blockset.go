package domain

type BlockSets []BlockSet

// BlockSet is a saved run of blocks the admin can drop into a case study as a
// starting point, the "header" set being the seeded one. It stores the same
// verbatim blocks a CaseLog does; there is no lifecycle, just name and blocks.
type BlockSet struct {
	Id     string `json:"id" bson:"_id,omitempty"`
	Name   string `json:"name" bson:"name,omitempty"`
	Blocks Blocks `json:"blocks" bson:"blocks,omitempty"`
}
