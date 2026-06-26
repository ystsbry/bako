package store

import (
	"bytes"
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"
)

// fmDelim is the line that opens and closes a YAML frontmatter block.
const fmDelim = "---"

// splitFrontmatter separates a markdown document into its YAML frontmatter and
// body. A document without a leading "---" line is treated as all body with
// empty frontmatter, so plain markdown files round-trip safely.
func splitFrontmatter(data []byte) (front []byte, body string) {
	s := string(bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n")))
	if !strings.HasPrefix(s, fmDelim+"\n") && s != fmDelim {
		return nil, s
	}
	rest := strings.TrimPrefix(s, fmDelim+"\n")
	// Find the closing delimiter at the start of a line.
	idx := strings.Index(rest, "\n"+fmDelim)
	if idx < 0 {
		// No closing delimiter; treat the whole thing as body.
		return nil, s
	}
	front = []byte(rest[:idx])
	after := rest[idx+len("\n"+fmDelim):]
	after = strings.TrimPrefix(after, "\n") // newline ending the closing "---" line
	after = strings.TrimPrefix(after, "\n") // blank separator line, when present
	return front, after
}

// parseDoc parses a markdown document into the given frontmatter struct and
// returns the trailing body. meta may be nil to skip YAML decoding.
func parseDoc(data []byte, meta any) (body string, err error) {
	front, body := splitFrontmatter(data)
	if meta != nil && len(front) > 0 {
		if err := yaml.Unmarshal(front, meta); err != nil {
			return "", fmt.Errorf("parse frontmatter: %w", err)
		}
	}
	return body, nil
}

// renderDoc serialises meta as YAML frontmatter followed by body. The body is
// emitted verbatim with exactly one trailing newline.
func renderDoc(meta any, body string) ([]byte, error) {
	var buf bytes.Buffer
	front, err := yaml.Marshal(meta)
	if err != nil {
		return nil, fmt.Errorf("render frontmatter: %w", err)
	}
	buf.WriteString(fmDelim + "\n")
	buf.Write(front)
	buf.WriteString(fmDelim + "\n\n")
	buf.WriteString(strings.TrimRight(body, "\n"))
	buf.WriteString("\n")
	return buf.Bytes(), nil
}
