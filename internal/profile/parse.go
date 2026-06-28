package profile

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// keyPattern is the set of valid environment variable names. Anything else is
// rejected at parse time so that emitted shell code cannot be subverted via a
// crafted key.
var keyPattern = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

// Parse reads a profile from path. Name is taken from the caller (derived from
// the filename) so this stays agnostic of the directory layout.
func Parse(path, name string) (*Profile, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	p := &Profile{Name: name, Path: path}
	// seen holds already-parsed literal values, used as the first lookup source
	// for $VAR expansion (process env is the fallback). Refs are not added.
	seen := map[string]string{}

	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
	lineNo := 0
	for sc.Scan() {
		lineNo++
		line := strings.TrimRight(sc.Text(), " \t\r")
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}
		if strings.HasPrefix(trimmed, "#") {
			if p.Desc == "" {
				if d, ok := parseDesc(trimmed); ok {
					p.Desc = d
				}
			}
			continue
		}

		// Optional leading "export ".
		assign := strings.TrimPrefix(trimmed, "export ")
		assign = strings.TrimSpace(assign)

		eq := strings.IndexByte(assign, '=')
		if eq < 0 {
			return nil, fmt.Errorf("%s:%d: not a KEY=VALUE assignment: %q", path, lineNo, line)
		}
		key := strings.TrimSpace(assign[:eq])
		if !keyPattern.MatchString(key) {
			return nil, fmt.Errorf("%s:%d: invalid variable name %q", path, lineNo, key)
		}

		rawVal := assign[eq+1:]
		val, quoted := stripQuotes(rawVal)

		if isRefValue(val) {
			p.Vars = append(p.Vars, Var{Key: key, Raw: val, IsRef: true})
			continue
		}

		// Single-quoted values are literal; double-quoted and unquoted values
		// undergo $VAR / ${VAR} expansion (shell semantics).
		if quoted != '\'' {
			val = expand(val, seen)
		}
		seen[key] = val
		p.Vars = append(p.Vars, Var{Key: key, Raw: val, IsRef: false})
	}
	if err := sc.Err(); err != nil {
		return nil, err
	}
	return p, nil
}

// parseDesc extracts the description from a "# desc: ..." comment.
func parseDesc(comment string) (string, bool) {
	body := strings.TrimLeft(comment, "#")
	body = strings.TrimSpace(body)
	const tag = "desc:"
	if !strings.HasPrefix(strings.ToLower(body), tag) {
		return "", false
	}
	return strings.TrimSpace(body[len(tag):]), true
}

// stripQuotes removes one matching pair of surrounding single or double quotes,
// returning the inner text and the quote rune used (0 if unquoted).
func stripQuotes(s string) (string, rune) {
	if len(s) >= 2 {
		c := s[0]
		if (c == '\'' || c == '"') && s[len(s)-1] == c {
			return s[1 : len(s)-1], rune(c)
		}
	}
	return s, 0
}

// expand resolves $VAR and ${VAR} against earlier literals first, then the
// process environment. Unknown variables expand to empty (shell behaviour).
func expand(s string, seen map[string]string) string {
	return os.Expand(s, func(name string) string {
		if v, ok := seen[name]; ok {
			return v
		}
		return os.Getenv(name)
	})
}
