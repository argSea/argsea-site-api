package out_port

import "github.com/argSea/argsea-site-api/argHex/domain"

// CaseStudySource feeds the caselog boot migration: it yields every project
// that still carries a non-empty legacy caseStudy string, with the header-seed
// fields alongside it. It is a read-only seam over the projects collection, the
// one place the retired caseStudy field is still decoded.
type CaseStudySource interface {
	LegacyCaseStudies() (domain.LegacyCaseStudies, error)
}
