package out_adapter

import (
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// caseStudySourceFake serves a fixed set of legacy case studies, standing in for
// the raw project-collection read in the migration tests.
type caseStudySourceFake struct {
	studies domain.LegacyCaseStudies
}

func NewCaseStudySourceFake(studies domain.LegacyCaseStudies) out_port.CaseStudySource {
	return caseStudySourceFake{
		studies: studies,
	}
}

func (c caseStudySourceFake) LegacyCaseStudies() (domain.LegacyCaseStudies, error) {
	return c.studies, nil
}
