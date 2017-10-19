package isos

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/koki/shorthand/ast"
)

// ServicePortsIso doot.
func ServicePortsIso() *ast.Iso {
	return ast.ZoomIso(ast.MkP(ast.P{"kind": "Service", "spec": ast.P{"ports": ast.AnyW}}),
		ast.MultiplyIso(Port()))
}

// Port doot.
func Port() *ast.Iso {
	from := ast.MkP(ast.XP{
		"name":     ast.StringW,
		"port":     ast.FloatW,
		"protocol": ast.StringW})
	to := ast.ConstPattern(ast.StringW)

	split := func(from *ast.Pattern) (*ast.Pattern, error) {
		x := from.Extract()

		name := ast.StringAt(x, "name")
		port := ast.FloatAt(x, "port")
		protocol := ast.StringAt(x, "protocol")

		if protocol == "TCP" {
			if name == "http" && port == 80 {
				return ast.ConstPattern("http"), nil
			}

			if name == "https" && port == 443 {
				return ast.ConstPattern("https"), nil
			}

			return ast.ConstPattern(fmt.Sprintf("%s:%v", name, port)), nil
		}

		return ast.ConstPattern(fmt.Sprintf("%s:%v:%s", name, port, protocol)), nil
	}

	unsplit := func(to *ast.Pattern) (*ast.Pattern, error) {
		x := to.ExtractString()
		segments := strings.Split(x, ":")
		l := len(segments)

		var err error
		var name string
		var port float64
		var protocol string

		if l > 0 {
			name = segments[0]
		} else {
			return nil, fmt.Errorf("expected non-empty string")
		}

		if l > 1 {
			port, err = strconv.ParseFloat(segments[1], 64)
			if err != nil {
				return nil, fmt.Errorf("couldn't parse port (%v)", err)
			}
		} else {
			switch name {
			case "http":
				port = 80
			case "https":
				port = 443
			default:
				return nil, fmt.Errorf("dunno port for %s", name)
			}
		}

		if l > 2 {
			protocol = segments[2]
		} else {
			protocol = "TCP"
		}

		return ast.MkP(ast.P{"name": name, "port": port, "protocol": protocol}), nil
	}

	return ast.MkIso(from, to, split, unsplit)
}
