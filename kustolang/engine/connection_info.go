package kustolang

import "strings"

// Ported from src/Kusto.Language/Utils/ConnectionInfo.cs.
// See third_party/kusto/LICENSE for the preserved Apache-2.0 license.

type ConnectionInfo struct {
	text  string
	parts map[string]string
}

func ParseConnectionInfo(text string) *ConnectionInfo {
	parts := map[string]string{}
	pos := 0
	for pos < len(text) {
		nameStart := pos
		for pos < len(text) && text[pos] != '=' && text[pos] != ';' {
			pos++
		}

		if pos < len(text) && text[pos] == '=' {
			name := text[nameStart:pos]
			pos++

			pos += scanConnectionWhitespace(text, pos)
			valueStart := pos

			if pos < len(text) && (text[pos] == '"' || text[pos] == '\'') {
				quote := text[pos]
				pos++
				for pos < len(text) && text[pos] != quote {
					pos++
				}
				if pos < len(text) && text[pos] == quote {
					pos++
					parts[connectionPartKey(name)] = text[valueStart+1 : pos-1]
				} else {
					parts[connectionPartKey(name)] = text[valueStart+1 : pos]
				}
			} else {
				for pos < len(text) && text[pos] != ';' {
					pos++
				}
				parts[connectionPartKey(name)] = strings.TrimSpace(text[valueStart:pos])
			}

			pos += scanConnectionWhitespace(text, pos)
		} else {
			value := strings.TrimSpace(text[nameStart:pos])
			parts[connectionPartKey("Data Source")] = value
		}

		if pos < len(text) && text[pos] == ';' {
			pos++
			continue
		}
		break
	}

	return &ConnectionInfo{text: text, parts: parts}
}

func (c *ConnectionInfo) GetPart(name string) string {
	if c == nil {
		return ""
	}
	return c.parts[connectionPartKey(name)]
}

func (c *ConnectionInfo) Text() string {
	if c == nil {
		return ""
	}
	return c.text
}

func (c *ConnectionInfo) DataSource() string {
	return c.GetPart("Data Source")
}

func connectionPartKey(name string) string {
	return ordinalIgnoreCaseKey(name)
}

func scanConnectionWhitespace(text string, start int) int {
	return ScanWhitespace(text, start)
}
