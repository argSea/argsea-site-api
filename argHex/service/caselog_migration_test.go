package service_test

import (
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
	"github.com/argSea/argsea-site-api/argHex/out_adapter"
	"github.com/argSea/argsea-site-api/argHex/service"
)

// oneLegacyStudy is a single project's dormant caseStudy plus the header inputs,
// enough to drive the migration end to end.
func oneLegacyStudy() domain.LegacyCaseStudy {
	return domain.LegacyCaseStudy{
		ProjectId: "proj-1",
		Title:     "Lighthouse",
		ShortDesc: "a small light",
		FirstLit:  "2019",
		Tags:      []string{"go", "mongo"},
		Facts:     []domain.ProjectFact{{Heading: "OWNERSHIP", Fact: "solo"}},
		CaseStudy: "# Story\n\nthe body text",
	}
}

func migBlockStr(b domain.Block, key string) string {
	s, _ := b[key].(string)
	return s
}

func TestMigrationPrependsTheHeaderThenTheParsedBody(t *testing.T) {
	repo := out_adapter.NewCaseLogFakeOutAdapter()
	migration := service.NewCaseLogMigration(repo, out_adapter.NewCaseStudySourceFake(domain.LegacyCaseStudies{oneLegacyStudy()}))

	n, err := migration.Run()

	if nil != err {
		t.Fatalf("migration failed: %v", err)
	}

	if 1 != n {
		t.Fatalf("expected one log migrated, got %d", n)
	}

	logs, _ := repo.List(false, 0)

	if 1 != len(logs) {
		t.Fatalf("expected one caselog created, got %d", len(logs))
	}

	log := logs[0]

	if domain.StatusPublished != log.Status {
		t.Fatalf("migrated logs must land published, got %q", log.Status)
	}

	if "proj-1" != log.ProjectId || "Lighthouse" != log.Title {
		t.Fatalf("expected the log keyed and titled from the project, got %+v", log)
	}

	// header first: title, subhead, facts, meta; then the parsed heading + paragraph
	if 6 != len(log.Blocks) {
		t.Fatalf("expected 4 header blocks + 2 body blocks, got %d", len(log.Blocks))
	}

	if "title" != migBlockStr(log.Blocks[0], "kind") || "Lighthouse" != migBlockStr(log.Blocks[0], "text") {
		t.Fatalf("block 0 should be the title, got %+v", log.Blocks[0])
	}

	if "subhead" != migBlockStr(log.Blocks[1], "kind") || "a small light" != migBlockStr(log.Blocks[1], "text") {
		t.Fatalf("block 1 should be the subhead, got %+v", log.Blocks[1])
	}

	factRows, ok := log.Blocks[2]["rows"].([]map[string]interface{})

	if "facts" != migBlockStr(log.Blocks[2], "kind") || !ok || 1 != len(factRows) || "OWNERSHIP" != factRows[0]["heading"] {
		t.Fatalf("block 2 should be the facts from the project, got %+v", log.Blocks[2])
	}

	tags, ok := log.Blocks[3]["tags"].([]string)

	if "meta" != migBlockStr(log.Blocks[3], "kind") || "2019" != migBlockStr(log.Blocks[3], "established") || !ok || 2 != len(tags) {
		t.Fatalf("block 3 should be the meta with established + tags, got %+v", log.Blocks[3])
	}

	if "heading" != migBlockStr(log.Blocks[4], "kind") || "Story" != migBlockStr(log.Blocks[4], "text") {
		t.Fatalf("block 4 should be the parsed heading, got %+v", log.Blocks[4])
	}

	if "paragraph" != migBlockStr(log.Blocks[5], "kind") || "the body text" != migBlockStr(log.Blocks[5], "text") {
		t.Fatalf("block 5 should be the parsed paragraph, got %+v", log.Blocks[5])
	}
}

func TestMigrationIsIdempotent(t *testing.T) {
	repo := out_adapter.NewCaseLogFakeOutAdapter()
	source := out_adapter.NewCaseStudySourceFake(domain.LegacyCaseStudies{oneLegacyStudy()})
	migration := service.NewCaseLogMigration(repo, source)

	if n, _ := migration.Run(); 1 != n {
		t.Fatalf("expected the first run to migrate 1, got %d", n)
	}

	// a second boot sees the same source (the legacy field is never unset) but
	// the project already owns a log, so nothing moves
	if n, _ := migration.Run(); 0 != n {
		t.Fatalf("expected the second run to migrate 0, got %d", n)
	}

	logs, _ := repo.List(false, 0)

	if 1 != len(logs) {
		t.Fatalf("expected still one log after two runs, got %d", len(logs))
	}
}

func TestMigrationSkipsProjectsThatAlreadyOwnALog(t *testing.T) {
	repo := out_adapter.NewCaseLogFakeOutAdapter()

	// proj-1 already has a log; only proj-2 should be migrated
	repo.Add(domain.CaseLog{ProjectId: "proj-1", Status: domain.StatusPublished, Title: "Existing"})

	source := out_adapter.NewCaseStudySourceFake(domain.LegacyCaseStudies{
		oneLegacyStudy(),
		{ProjectId: "proj-2", Title: "Second", CaseStudy: "just a paragraph"},
	})

	n, err := service.NewCaseLogMigration(repo, source).Run()

	if nil != err {
		t.Fatalf("migration failed: %v", err)
	}

	if 1 != n {
		t.Fatalf("expected only the un-logged project migrated, got %d", n)
	}

	logs, _ := repo.List(false, 0)

	if 2 != len(logs) {
		t.Fatalf("expected the pre-existing log plus one new one, got %d", len(logs))
	}
}
