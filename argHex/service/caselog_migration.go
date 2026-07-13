package service

import (
	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_port"
)

// CaseLogMigration lifts every project's dormant caseStudy into its own
// published caselog, once, at boot. It crosses two collections (reads projects
// through a CaseStudySource, writes caselogs), so it is its own unit rather than
// a repo method like the hobby ships-log migration; it follows the same shape
// otherwise: idempotent, run before the routes mount, one log line.
type CaseLogMigration struct {
	logs   out_port.CaseLogRepo
	source out_port.CaseStudySource
}

func NewCaseLogMigration(logs out_port.CaseLogRepo, source out_port.CaseStudySource) *CaseLogMigration {
	return &CaseLogMigration{
		logs:   logs,
		source: source,
	}
}

// Run creates a published caselog for every project that carries a legacy
// caseStudy and does not already own one, seeding the header from the project's
// own fields ahead of the parsed body. Idempotent: a project whose log already
// exists is skipped, so a reboot creates nothing. Returns the count created.
func (m *CaseLogMigration) Run() (int, error) {
	sources, err := m.source.LegacyCaseStudies()

	if nil != err {
		return 0, err
	}

	existing, err := m.logs.List(false, 0)

	if nil != err {
		return 0, err
	}

	claimed := map[string]bool{}

	for _, log := range existing {
		claimed[log.ProjectId] = true
	}

	migrated := 0

	for _, src := range sources {
		if claimed[src.ProjectId] {
			continue
		}

		now := nowStamp()
		log := domain.CaseLog{
			ProjectId:   src.ProjectId,
			Status:      domain.StatusPublished,
			Title:       src.Title,
			Blocks:      append(headerBlocks(src), parseDialect(src.CaseStudy)...),
			PublishedAt: now,
			CreatedAt:   now,
			UpdatedAt:   now,
		}

		if _, err := m.logs.Add(log); nil != err {
			return migrated, err
		}

		claimed[src.ProjectId] = true
		migrated++
	}

	return migrated, nil
}

// headerBlocks prepends the header the site used to derive from the project doc:
// title from the project title, subhead from its short description, a facts
// block from its facts, and a meta block carrying the first-lit year and tags.
func headerBlocks(src domain.LegacyCaseStudy) domain.Blocks {
	factRows := make([]map[string]interface{}, 0, len(src.Facts))

	for _, fact := range src.Facts {
		factRows = append(factRows, map[string]interface{}{"heading": fact.Heading, "fact": fact.Fact})
	}

	tags := src.Tags

	if nil == tags {
		tags = []string{}
	}

	return domain.Blocks{
		domain.Block{"kind": "title", "text": src.Title},
		domain.Block{"kind": "subhead", "text": src.ShortDesc},
		domain.Block{"kind": "facts", "rows": factRows},
		domain.Block{"kind": "meta", "established": src.FirstLit, "tags": tags},
	}
}
