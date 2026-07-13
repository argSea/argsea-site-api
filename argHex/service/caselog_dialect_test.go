package service

import (
	"testing"

	"github.com/argSea/argsea-site-api/argHex/domain"
)

// The parser is the migration's port of caseStudy.ts scanBlocks, so the cases
// mirror the grammar the mock defined: headings, paragraph/quote runs, lists,
// fenced code, :::facts/:::outcomes, and figures, with inline marks left verbatim.

func blockStr(b domain.Block, key string) string {
	s, _ := b[key].(string)
	return s
}

func TestParseHeadingLevelsStripTheMarker(t *testing.T) {
	blocks := parseDialect("# One\n## Two\n### Three")

	if 3 != len(blocks) {
		t.Fatalf("expected three heading blocks, got %d", len(blocks))
	}

	for i, want := range []string{"One", "Two", "Three"} {
		if "heading" != blockStr(blocks[i], "kind") || want != blockStr(blocks[i], "text") {
			t.Fatalf("heading %d was %+v", i, blocks[i])
		}
	}
}

func TestParseParagraphRunJoinsWithASpace(t *testing.T) {
	blocks := parseDialect("alpha\nbeta\ngamma")

	if 1 != len(blocks) || "paragraph" != blockStr(blocks[0], "kind") {
		t.Fatalf("expected one paragraph, got %+v", blocks)
	}

	if "alpha beta gamma" != blockStr(blocks[0], "text") {
		t.Fatalf("expected the run joined with spaces, got %q", blockStr(blocks[0], "text"))
	}
}

func TestParseQuoteRunStripsMarkerAndJoins(t *testing.T) {
	blocks := parseDialect("> from the log\n> the second line")

	if 1 != len(blocks) || "quote" != blockStr(blocks[0], "kind") {
		t.Fatalf("expected one quote, got %+v", blocks)
	}

	if "from the log the second line" != blockStr(blocks[0], "text") {
		t.Fatalf("expected the quote joined, got %q", blockStr(blocks[0], "text"))
	}
}

func TestParseListTakesBothMarkersUnordered(t *testing.T) {
	blocks := parseDialect("- first\n* second")

	if 1 != len(blocks) || "list" != blockStr(blocks[0], "kind") {
		t.Fatalf("expected one list, got %+v", blocks)
	}

	if false != blocks[0]["ordered"] {
		t.Fatalf("the dialect only makes unordered lists, got ordered %v", blocks[0]["ordered"])
	}

	items, ok := blocks[0]["items"].([]string)

	if !ok || 2 != len(items) || "first" != items[0] || "second" != items[1] {
		t.Fatalf("expected both items with the marker stripped, got %v", blocks[0]["items"])
	}
}

func TestParseCodeFenceKeepsLangAndMermaidSplitsOff(t *testing.T) {
	code := parseDialect("```go\nfmt.Println(\"hi\")\n```")

	if 1 != len(code) || "code" != blockStr(code[0], "kind") {
		t.Fatalf("expected one code block, got %+v", code)
	}

	if "go" != blockStr(code[0], "lang") || "fmt.Println(\"hi\")" != blockStr(code[0], "code") {
		t.Fatalf("expected the lang kept and body verbatim, got %+v", code[0])
	}

	mermaid := parseDialect("```mermaid\ngraph TD\nA-->B\n```")

	if 1 != len(mermaid) || "mermaid" != blockStr(mermaid[0], "kind") {
		t.Fatalf("expected a mermaid block, got %+v", mermaid)
	}

	if "graph TD\nA-->B" != blockStr(mermaid[0], "code") {
		t.Fatalf("expected the mermaid body verbatim, got %q", blockStr(mermaid[0], "code"))
	}
}

func TestParseFactsSplitOnFirstColon(t *testing.T) {
	blocks := parseDialect(":::facts\nFounded: 2019\nno colon here\n:::")

	if 1 != len(blocks) || "facts" != blockStr(blocks[0], "kind") {
		t.Fatalf("expected a facts block, got %+v", blocks)
	}

	rows, ok := blocks[0]["rows"].([]map[string]interface{})

	if !ok || 2 != len(rows) {
		t.Fatalf("expected two fact rows, got %v", blocks[0]["rows"])
	}

	if "Founded" != rows[0]["heading"] || "2019" != rows[0]["fact"] {
		t.Fatalf("expected the first colon to split heading/fact, got %+v", rows[0])
	}

	// a colon-less line is an unlabeled fact
	if "" != rows[1]["heading"] || "no colon here" != rows[1]["fact"] {
		t.Fatalf("expected a colon-less row unlabeled, got %+v", rows[1])
	}
}

func TestParseOutcomesSplitOnPipeAndDropThirdSegment(t *testing.T) {
	blocks := parseDialect(":::outcomes\n99.9% | uptime | dropped\n50ms | latency\n:::")

	if 1 != len(blocks) || "outcomes" != blockStr(blocks[0], "kind") {
		t.Fatalf("expected an outcomes block, got %+v", blocks)
	}

	rows, ok := blocks[0]["rows"].([]map[string]interface{})

	if !ok || 2 != len(rows) {
		t.Fatalf("expected two outcome rows, got %v", blocks[0]["rows"])
	}

	// the third pipe segment is dropped, same as the mock
	if "99.9%" != rows[0]["value"] || "uptime" != rows[0]["caption"] {
		t.Fatalf("expected value/caption with the third dropped, got %+v", rows[0])
	}

	if "50ms" != rows[1]["value"] || "latency" != rows[1]["caption"] {
		t.Fatalf("expected the second outcome split, got %+v", rows[1])
	}
}

func TestParseFigureCapturesImageAndCaption(t *testing.T) {
	blocks := parseDialect("![A wide shot](harbor.jpg)")

	if 1 != len(blocks) || "figure" != blockStr(blocks[0], "kind") {
		t.Fatalf("expected a figure block, got %+v", blocks)
	}

	if "harbor.jpg" != blockStr(blocks[0], "image") || "A wide shot" != blockStr(blocks[0], "caption") {
		t.Fatalf("expected image from the target and caption from the alt, got %+v", blocks[0])
	}
}

func TestParseLeavesInlineMarksVerbatim(t *testing.T) {
	blocks := parseDialect("Some **bold** and `code` and [? chip ?] stay put")

	if 1 != len(blocks) || "paragraph" != blockStr(blocks[0], "kind") {
		t.Fatalf("expected a paragraph, got %+v", blocks)
	}

	// the API does not parse inline marks; only the site does
	if "Some **bold** and `code` and [? chip ?] stay put" != blockStr(blocks[0], "text") {
		t.Fatalf("expected the inline syntax untouched, got %q", blockStr(blocks[0], "text"))
	}
}
