package isos

import (
	"github.com/koki/shorthand/ast"
)

// ManifestIso shorthand format for an entire k8s manifest.
func ManifestIso() *ast.Iso {
	return ast.MultiplyIso(ast.SequenceIsos(ServicePortsIso(), ast.IdentityIso()))
}
