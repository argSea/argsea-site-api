package service

import (
	"regexp"
	"strings"

	"github.com/argSea/argsea-site-api/argHex/domain"
)

// The dialect parser ports argsea-site/src/lib/caseStudy.ts scanBlocks: the same
// narrow line-scanner the site built case studies with, mapped onto the block
// union. Only the legacy grammar is ported; the new block kinds never had a
// dialect, so nothing here emits them. Inline marks (**bold**, `code`,
// [? chips ?], links) stay in the text verbatim, since the site is the only
// place that parses them.

var dialectHeading = regexp.MustCompile(`^#{1,3} `)
var dialectListItem = regexp.MustCompile(`^[-*] `)
var dialectFigure = regexp.MustCompile(`^!\[([^\]]*)\]\(([^)\s]+)\)\s*$`)

// dialectStop is the paragraph-continuation guard: a paragraph run swallows
// following lines until a blank line or one of these openers, mirroring the
// mock's regex exactly (the img opener is deliberately absent from it, same as
// the mock).
var dialectStop = regexp.MustCompile("^(#{1,3} |> |:::|```|[-*] )")

// parseDialect scans a legacy case study string into verbatim blocks, mirroring
// caseStudy.ts scanBlocks: fenced code (mermaid split off, other langs kept),
// :::facts/:::outcomes fences, #/##/### headings, ![caption](image) figures,
// > quote runs, -/* list runs, and paragraphs for everything else.
func parseDialect(md string) domain.Blocks {
	lines := strings.Split(md, "\n")
	blocks := domain.Blocks{}
	i := 0

	for i < len(lines) {
		line := lines[i]

		if "" == strings.TrimSpace(line) {
			i++
			continue
		}

		if strings.HasPrefix(line, "```") {
			lang := strings.ToLower(strings.TrimSpace(line[3:]))
			var buf []string
			i++

			for i < len(lines) && !strings.HasPrefix(lines[i], "```") {
				buf = append(buf, lines[i])
				i++
			}

			i++
			body := strings.Join(buf, "\n")

			if "mermaid" == lang {
				blocks = append(blocks, domain.Block{"kind": "mermaid", "code": body})
			} else {
				blocks = append(blocks, domain.Block{"kind": "code", "lang": lang, "code": body})
			}

			continue
		}

		if strings.HasPrefix(line, ":::") {
			kind := strings.ToLower(strings.TrimSpace(line[3:]))
			var buf []string
			i++

			for i < len(lines) && !strings.HasPrefix(lines[i], ":::") {
				if "" != strings.TrimSpace(lines[i]) {
					buf = append(buf, strings.TrimSpace(lines[i]))
				}
				i++
			}

			i++

			if "outcomes" == kind {
				blocks = append(blocks, domain.Block{"kind": "outcomes", "rows": dialectOutcomeRows(buf)})
			} else {
				blocks = append(blocks, domain.Block{"kind": "facts", "rows": dialectFactRows(buf)})
			}

			continue
		}

		if dialectHeading.MatchString(line) {
			text := strings.TrimSpace(dialectHeading.ReplaceAllString(line, ""))
			blocks = append(blocks, domain.Block{"kind": "heading", "text": text})
			i++
			continue
		}

		if m := dialectFigure.FindStringSubmatch(line); nil != m {
			blocks = append(blocks, domain.Block{"kind": "figure", "image": m[2], "caption": m[1]})
			i++
			continue
		}

		if strings.HasPrefix(line, "> ") {
			buf := []string{line[2:]}
			i++

			for i < len(lines) && strings.HasPrefix(lines[i], "> ") {
				buf = append(buf, lines[i][2:])
				i++
			}

			blocks = append(blocks, domain.Block{"kind": "quote", "text": strings.Join(buf, " ")})
			continue
		}

		if dialectListItem.MatchString(line) {
			items := []string{}

			for i < len(lines) && dialectListItem.MatchString(lines[i]) {
				items = append(items, lines[i][2:])
				i++
			}

			blocks = append(blocks, domain.Block{"kind": "list", "ordered": false, "items": items})
			continue
		}

		buf := []string{line}
		i++

		for i < len(lines) && "" != strings.TrimSpace(lines[i]) && !dialectStop.MatchString(lines[i]) {
			buf = append(buf, lines[i])
			i++
		}

		blocks = append(blocks, domain.Block{"kind": "paragraph", "text": strings.Join(buf, " ")})
	}

	return blocks
}

// dialectFactRows splits each :::facts line on its FIRST colon into a
// heading/fact pair; a colon-less line is an unlabeled fact, verbatim to the
// mock's splitLabelValue.
func dialectFactRows(rows []string) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))

	for _, row := range rows {
		heading, fact := splitFirstColon(row)
		out = append(out, map[string]interface{}{"heading": heading, "fact": fact})
	}

	return out
}

// dialectOutcomeRows splits each :::outcomes line on "|" into a value/caption
// pair; a third segment, if any, is dropped, verbatim to the mock's
// splitOutcome.
func dialectOutcomeRows(rows []string) []map[string]interface{} {
	out := make([]map[string]interface{}, 0, len(rows))

	for _, row := range rows {
		value, caption := splitOutcome(row)
		out = append(out, map[string]interface{}{"value": value, "caption": caption})
	}

	return out
}

func splitFirstColon(row string) (string, string) {
	colon := strings.Index(row, ":")

	if colon > -1 {
		return strings.TrimSpace(row[:colon]), strings.TrimSpace(row[colon+1:])
	}

	return "", row
}

func splitOutcome(row string) (string, string) {
	parts := strings.Split(row, "|")
	value := ""
	caption := ""

	if 0 < len(parts) {
		value = strings.TrimSpace(parts[0])
	}

	if 1 < len(parts) {
		caption = strings.TrimSpace(parts[1])
	}

	return value, caption
}
