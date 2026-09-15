package main

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

// manifest mirrors bilc.DeployManifest's JSON shape (see
// ../../../bil/tools/bilc/bilc.go's DeployManifest doc comment) -- this is
// the same on-disk file static/index.html's resolveRole reads for the
// browser backend.
type manifest struct {
	Transport string `json:"transport"`
	Leaves    []leaf `json:"leaves"`
}

type leaf struct {
	Match      matchClause `json:"match"`
	Conditions []condition `json:"conditions"`
	Proc       string      `json:"proc"`
}

type matchClause struct {
	ID      string `json:"id"`
	Row     string `json:"row"`
	Col     string `json:"col"`
	Default bool   `json:"default"`
}

type condition struct {
	Expr   string `json:"expr"`
	Negate bool   `json:"negate"`
}

func loadManifest(path string) (*manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var m manifest
	if err := json.Unmarshal(data, &m); err != nil {
		return nil, fmt.Errorf("parsing %s: %w", path, err)
	}
	return &m, nil
}

// resolveRole is a direct Go port of static/index.html's resolveRole: walk
// leaves in order (first match wins, matching bilc's own switch/if-else
// semantics) and return the proc that should run at (r,c).
func resolveRole(m *manifest, r, c, rows, cols int) (string, error) {
	vars := map[string]int{"r": r, "c": c, "rows": rows, "cols": cols}
	for _, l := range m.Leaves {
		matched, err := clauseMatches(l.Match, vars, r, c)
		if err != nil {
			return "", err
		}
		if !matched {
			continue
		}
		ok := true
		for _, cond := range l.Conditions {
			v, err := evalBool(cond.Expr, vars)
			if err != nil {
				return "", err
			}
			if cond.Negate {
				v = !v
			}
			if !v {
				ok = false
				break
			}
		}
		if ok {
			return l.Proc, nil
		}
	}
	return "", fmt.Errorf("no placed-par leaf matches (r=%d, c=%d)", r, c)
}

func clauseMatches(m matchClause, vars map[string]int, r, c int) (bool, error) {
	switch {
	case m.Default:
		return true, nil
	case m.ID != "":
		v, err := evalInt(m.ID, vars)
		if err != nil {
			return false, err
		}
		return v == r*vars["cols"]+c, nil
	default:
		rowOK := true
		if m.Row != "*" {
			v, err := evalInt(m.Row, vars)
			if err != nil {
				return false, err
			}
			rowOK = v == r
		}
		colOK := true
		if m.Col != "*" {
			v, err := evalInt(m.Col, vars)
			if err != nil {
				return false, err
			}
			colOK = v == c
		}
		return rowOK && colOK, nil
	}
}

func evalBool(src string, vars map[string]int) (bool, error) {
	v, err := evalExpr(src, vars)
	if err != nil {
		return false, err
	}
	switch v := v.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	default:
		return false, fmt.Errorf("evalExpr: %q did not evaluate to a boolean", src)
	}
}

func evalInt(src string, vars map[string]int) (int, error) {
	v, err := evalExpr(src, vars)
	if err != nil {
		return 0, err
	}
	switch v := v.(type) {
	case int:
		return v, nil
	case bool:
		return 0, fmt.Errorf("evalExpr: %q evaluated to a boolean, want an int", src)
	default:
		return 0, fmt.Errorf("evalExpr: %q did not evaluate to an int", src)
	}
}

// evalExpr is a direct Go port of static/index.html's evalExpr: the tiny
// grammar bilc's placed-par clause-match/leaf-condition expressions
// actually use -- identifiers (r, c, rows, cols), int literals, ==, !=,
// &&, +, -. Returns int or bool (mirroring the JS original's dynamic
// typing), matching whichever an expression actually reduces to. See
// index.html's own comment for why this is hand-rolled rather than
// reaching for a general expression evaluator.
func evalExpr(src string, vars map[string]int) (any, error) {
	p := &exprParser{src: src, vars: vars}
	v, err := p.parseAnd()
	if err != nil {
		return nil, err
	}
	p.skipWs()
	if p.pos != len(p.src) {
		return nil, fmt.Errorf("evalExpr: trailing input in %q", p.src)
	}
	return v, nil
}

type exprParser struct {
	src  string
	pos  int
	vars map[string]int
}

var identRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*`)
var numRe = regexp.MustCompile(`^[0-9]+`)

func (p *exprParser) skipWs() {
	for p.pos < len(p.src) && (p.src[p.pos] == ' ' || p.src[p.pos] == '\t') {
		p.pos++
	}
}

func (p *exprParser) consume(tok string) bool {
	p.skipWs()
	if strings.HasPrefix(p.src[p.pos:], tok) {
		p.pos += len(tok)
		return true
	}
	return false
}

func asInt(v any, src string) (int, error) {
	n, ok := v.(int)
	if !ok {
		return 0, fmt.Errorf("evalExpr: expected an int in %q", src)
	}
	return n, nil
}

func (p *exprParser) parsePrimary() (any, error) {
	p.skipWs()
	if p.consume("(") {
		v, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		if !p.consume(")") {
			return nil, fmt.Errorf("evalExpr: expected ) in %q", p.src)
		}
		return v, nil
	}
	if m := identRe.FindString(p.src[p.pos:]); m != "" {
		p.pos += len(m)
		v, ok := p.vars[m]
		if !ok {
			return nil, fmt.Errorf("evalExpr: unknown identifier %q in %q", m, p.src)
		}
		return v, nil
	}
	if m := numRe.FindString(p.src[p.pos:]); m != "" {
		p.pos += len(m)
		n, _ := strconv.Atoi(m)
		return n, nil
	}
	return nil, fmt.Errorf("evalExpr: unexpected token at %d in %q", p.pos, p.src)
}

func (p *exprParser) parseAdd() (any, error) {
	lhs, err := p.parsePrimary()
	if err != nil {
		return nil, err
	}
	v, err := asInt(lhs, p.src)
	if err != nil {
		return nil, err
	}
	for {
		if p.consume("+") {
			rhs, err := p.parsePrimary()
			if err != nil {
				return nil, err
			}
			r, err := asInt(rhs, p.src)
			if err != nil {
				return nil, err
			}
			v += r
		} else if p.consume("-") {
			rhs, err := p.parsePrimary()
			if err != nil {
				return nil, err
			}
			r, err := asInt(rhs, p.src)
			if err != nil {
				return nil, err
			}
			v -= r
		} else {
			return v, nil
		}
	}
}

func (p *exprParser) parseCmp() (any, error) {
	lhs, err := p.parseAdd()
	if err != nil {
		return nil, err
	}
	if p.consume("==") {
		rhs, err := p.parseAdd()
		if err != nil {
			return nil, err
		}
		return lhs == rhs, nil
	}
	if p.consume("!=") {
		rhs, err := p.parseAdd()
		if err != nil {
			return nil, err
		}
		return lhs != rhs, nil
	}
	return lhs, nil
}

func (p *exprParser) parseAnd() (any, error) {
	v, err := p.parseCmp()
	if err != nil {
		return nil, err
	}
	for p.consume("&&") {
		rhs, err := p.parseCmp()
		if err != nil {
			return nil, err
		}
		vb, err := evalBoolVal(v, p.src)
		if err != nil {
			return nil, err
		}
		rb, err := evalBoolVal(rhs, p.src)
		if err != nil {
			return nil, err
		}
		v = vb && rb
	}
	return v, nil
}

func evalBoolVal(v any, src string) (bool, error) {
	switch v := v.(type) {
	case bool:
		return v, nil
	case int:
		return v != 0, nil
	default:
		return false, fmt.Errorf("evalExpr: unsupported value in %q", src)
	}
}
