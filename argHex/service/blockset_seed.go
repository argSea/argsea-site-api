package service

import "github.com/argSea/argsea-site-api/argHex/domain"

// seedHeaderBlockSet is the one set planted at boot: an empty case study header
// the admin fills in. Title and subhead are blank, the facts carry the three
// standing headings with empty facts, and the meta block starts empty.
func seedHeaderBlockSet() domain.BlockSet {
	return domain.BlockSet{
		Name: "header",
		Blocks: domain.Blocks{
			domain.Block{"kind": "title", "text": ""},
			domain.Block{"kind": "subhead", "text": ""},
			domain.Block{"kind": "facts", "rows": []map[string]interface{}{
				{"heading": "OWNERSHIP", "fact": ""},
				{"heading": "OUTCOME", "fact": ""},
				{"heading": "SCOPE", "fact": ""},
			}},
			domain.Block{"kind": "meta", "established": "", "tags": []string{}},
		},
	}
}
