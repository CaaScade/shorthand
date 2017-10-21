package cmd

// This file contains methods for recursively inspecting a type definition
//   and a Print() method to test their implementation.

import (
	"fmt"
	"go/ast"
	"go/token"
	"go/types"
	"strings"

	"github.com/golang/glog"
	"github.com/kr/pretty"
	"golang.org/x/tools/go/loader"
	v1 "k8s.io/api/core/v1"
)

// load a package and traverse all its types.
func load() {
	// Just so the compiler doesn't complain about not using "v1".
	var dummy v1.Pod
	_ = dummy.Spec

	var conf loader.Config

	// We're just loading "v1" (and then all its dependencies).
	conf.Import("k8s.io/api/core/v1")
	program, err := conf.Load()

	// If we're missing dependencies or our "v1" is otherwise broken, quit.
	if err != nil {
		glog.Fatal(err)
	}

	// Create a traversal context for each type definition in "v1".
	contexts := ContextsForPackage(program, program.InitialPackages()[0])

	// Test the traversal context by printing all fields "recursively".
	for _, context := range contexts {
		context.Print(0)
	}

	fmt.Println("yes, hello")
}

// SourceField is the name of a source field and its type information.
type SourceField struct {
	// Name of the field in the source object.
	Name string
	// The type information of this field.
	// The important information here is whether the type needs a
	//   nil check before its subfields are accessed.
	TypeExpr ast.Expr
	// A Context focused on the struct this field is part of.
	// This is used to interpret TypeExpr properly.
	Context *Context
}

// MappedField contains either a new struct type or a value extracted from the
//   source object.
type MappedField struct {
	// Name is the name of the new field.
	Name string
	// NewStruct is nil if this MappedField is just mapped from a single value
	//   from the source object.
	// Otherwise, it's a whole new struct type.
	NewStruct *MappedStruct
	// OriginPath tells us all the field names and types from the root
	//   source object to the value we want to insert at Name in this
	//   destination object (which may not be the root).
	// Only used if NewStruct is nil.
	OriginPath []*SourceField
}

// MappedStruct is a new struct type created by moving around the fields of
//   a source struct.
type MappedStruct struct {
	// Name of the new struct type.
	Name string
	// The fields of this struct.
	Fields []*MappedField
}

// Context is a cursor that indicates our current position in the program.
// It contains the information needed to understand every part of TypeSpec.
type Context struct {
	// Program is the entire loaded program.
	Program *loader.Program
	// Package is an entire loaded package.
	Package *loader.PackageInfo
	// File is a file in the Package.
	// The file matters because the local name of an import is
	//   file-specific. This is how we resolve Selector expressions.
	//   e.g. metav1.ObjectMeta
	File *ast.File
	// TypeSpec is a type definition in the File.
	TypeSpec *ast.TypeSpec
}

// importMatchesPackage checks if a package corresponds to an import.
// types.Package uses the actual filesystem path (for non-built-in packages)
//   e.g. github.com/koki/shorthand/vendor/k8s.io/apimachinery
//        go/ast
// ast.ImportSpec uses the quoted string in the source file's import statement.
//   e.g. "k8s.io/apimachinery" <- with quotes
//        "go/ast"
func importMatchesPackage(imprt *ast.ImportSpec, pkg *types.Package) bool {
	quoted := imprt.Path.Value
	stripped := quoted[1 : len(quoted)-1]
	return pathMatchesPackage(stripped, pkg)
}

// pathMatchesPackage checks if a path corresponds to a given package.
// pkgPath can be either of two formats:
//   e.g. github.com/koki/shorthand/vendor/k8s.io/apimachinery (types.Package)
//   e.g. k8s.io/apimachinery (unquoted ast.ImportSpec)
func pathMatchesPackage(pkgPath string, pkg *types.Package) bool {
	// If pkgPath is the long format, then match this way.
	if pkgPath == pkg.Path() {
		return true
	}

	// If pkgPath is the short format, then match this way.
	if strings.HasSuffix(pkg.Path(), "vendor/"+pkgPath) {
		return true
	}

	return false
}

// ContextForImportedType constructs a context for an import
//   and the name of a type.
// It traverses the Program to find the right Package and TypeSpec.
func ContextForImportedType(program *loader.Program, imprt *ast.ImportSpec, typeName string) *Context {
	quoted := imprt.Path.Value
	stripped := quoted[1 : len(quoted)-1]
	return ContextForType(program, stripped, typeName)
}

// ContextForType constructs a context for a pkg path (see pathMatchesPackage)
//   and the name of a type.
// It traverses the Program to find the right Package and TypeSpec.
func ContextForType(program *loader.Program, pkgPath, typeName string) *Context {
	for _, pkg := range program.AllPackages {
		if !pathMatchesPackage(pkgPath, pkg.Pkg) {
			continue
		}

		context := ContextForPackageAndType(program, pkg, typeName)
		if context != nil {
			return context
		}
	}

	return nil
}

// ContextForPackageAndType constructs a context for a package and the name of a type.
// It traverses only the given package to find the right TypeSpec.
func ContextForPackageAndType(program *loader.Program, pkg *loader.PackageInfo, typeName string) *Context {
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				if decl.Tok == token.TYPE {
					for _, spec := range decl.Specs {
						switch spec := spec.(type) {
						case *ast.TypeSpec:
							if spec.Name.Name == typeName {
								return &Context{Program: program, Package: pkg, File: file, TypeSpec: spec}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

// ContextsForPackage traverses an entire package and creates a context for each
//   type definition it finds.
func ContextsForPackage(program *loader.Program, pkg *loader.PackageInfo) []*Context {
	contexts := []*Context{}
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				if decl.Tok == token.TYPE {
					for _, spec := range decl.Specs {
						switch spec := spec.(type) {
						case *ast.TypeSpec:
							contexts = append(contexts, &Context{Program: program, Package: pkg, File: file, TypeSpec: spec})
						}
					}
				}
			}
		}
	}

	return contexts
}

// RefocusedWithinPackage navigates from one type (TypeSpec) to another within
//   the same package.
// For example, this would be used to go from inspecting v1.Pod to v1.PodSpec
//   When we inspect v1.Pod, the Spec field gives us a TypeSpec object for
//   v1.PodSpec.
func (context *Context) RefocusedWithinPackage(typeSpec *ast.TypeSpec) *Context {
	for _, file := range context.Package.Files {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				if decl.Tok == token.TYPE {
					for _, spec := range decl.Specs {
						if typeSpec == spec {
							return &Context{Program: context.Program, Package: context.Package, File: file, TypeSpec: typeSpec}
						}
					}
				}
			}
		}
	}

	return nil
}

// RefocusedWithSelector navigates from one context to a context in a different package.
// For example, this would be used to go from v1.Pod to metav1.ObjectMeta.
//   When we inspect v1.Pod, an anonymous field gives us a Selector for metav1.ObjectMeta.
func (context *Context) RefocusedWithSelector(selector *ast.SelectorExpr) *Context {
	var pkgName string
	switch expr := selector.X.(type) {
	case *ast.Ident:
		pkgName = expr.Name
	default:
		glog.Fatal(pretty.Sprint(selector))
	}

	typeName := selector.Sel.Name

	return context.refocusedWithPkgAndTypeNames(pkgName, typeName)
}

// getImportedPackage traverses the Program to find the Package that matches
//   a given Import.
// It's a helper function for refocusedWithPkgAndTypeNames.
func (context *Context) getImportedPackage(imprt *ast.ImportSpec) *loader.PackageInfo {
	for _, pkg := range context.Program.AllPackages {
		if importMatchesPackage(imprt, pkg.Pkg) {
			return pkg
		}
	}

	return nil
}

// refocusedWithPkgAndTypeNames traverses the Program to find a package with
//   a given name (not the path, but the name used to prefix imported definitions)
// It's a helper function for RefocusedWithSelector.
func (context *Context) refocusedWithPkgAndTypeNames(pkgName string, typeName string) *Context {
	for _, imprt := range context.File.Imports {
		if imprt.Name != nil {
			// If the local name matches, look up the type here.
			if imprt.Name.Name == pkgName {
				return ContextForImportedType(context.Program, imprt, typeName)
			}
		} else {
			// If the default name matches, look up the type here.
			pkg := context.getImportedPackage(imprt)
			if pkg.Pkg.Name() == pkgName {
				return ContextForPackageAndType(context.Program, pkg, typeName)
			}
		}
	}

	return nil
}

// These (and init()) are just for formatting the Print methods.
var maxDepth = 40
var indents = make([]string, maxDepth)

func init() {
	for index := range indents {
		indents[index] = strings.Repeat("  ", index)
	}
}

// Print recursively traverses all the fields of a type and prints them.
// "depth" is the initial indentation depth.
func (context *Context) Print(depth int) {
	if context == nil {
		glog.Fatal()
	}

	fmt.Println(indents[depth], context.TypeSpec.Name)
	context.PrintType(depth+1, context.TypeSpec.Type)
}

// PrintType prints the contents (RHS) of a type declaration.
// "depth" is the initial indentation depth.
func (context *Context) PrintType(depth int, root ast.Expr) {
	switch root := root.(type) {
	case *ast.ParenExpr:
		// Strip parens
		context.PrintType(depth, root.X)
	case *ast.Ident:
		// Print name and then associated object.
		context.PrintTypeIdent(depth, root)
	case *ast.SelectorExpr:
		context.PrintSelector(depth, root)
	case *ast.StarExpr:
		fmt.Println(indents[depth], "*")
		context.PrintType(depth, root.X)
	case *ast.FuncType:
		fmt.Println(indents[depth], "func")
	case *ast.ChanType:
		fmt.Println(indents[depth], "chan")
	case *ast.ArrayType:
		fmt.Println(indents[depth], "[]")
		context.PrintType(depth, root.Elt)
	case *ast.StructType:
		for _, field := range root.Fields.List {
			context.PrintField(depth, field)
		}
	case *ast.MapType:
		fmt.Println(indents[depth], "map")
		context.PrintType(depth, root.Key)
		context.PrintType(depth, root.Value)
	case *ast.InterfaceType:
		fmt.Println(indents[depth], "interface")
	default:
		glog.Fatal(pretty.Printf("non-type expr (%# v)", root))
	}
}

// PrintSelector prints a Selector and the details of the type it represents.
// "depth" is the initial indentation depth.
func (context *Context) PrintSelector(depth int, root *ast.SelectorExpr) {
	var pkgName string
	switch expr := root.X.(type) {
	case *ast.Ident:
		pkgName = expr.Name
	default:
		glog.Fatal(pretty.Sprint(root))
	}

	typeName := root.Sel.Name

	fmt.Printf("%s%s.%s\n", indents[depth], pkgName, typeName)
	context.RefocusedWithSelector(root).Print(depth)
}

// PrintTypeIdent prints a type identifier along with its contents.
// An Ident either refers to a built-in type or a type within the same Package.
// "depth" is the initial indentation depth.
func (context *Context) PrintTypeIdent(depth int, root *ast.Ident) {
	obj := root.Obj
	if obj != nil {
		switch obj.Kind {
		case ast.Typ:
			switch decl := obj.Decl.(type) {
			case *ast.TypeSpec:
				context.RefocusedWithinPackage(decl).Print(depth)
			default:
				fmt.Println(indents[depth], root.Name)
				fmt.Println(indents[depth+1], "Typ but no TypeSpec")
			}

			return

		default:
			fmt.Println(indents[depth], root.Name)
			fmt.Println(indents[depth+1], obj)
		}

		return
	}

	fmt.Println(indents[depth], root.Name)
}

// PrintField prints a struct Field and its type information.
// "depth" is the initial indentation depth.
func (context *Context) PrintField(depth int, root *ast.Field) {
	l := len(root.Names)
	if l > 0 {
		names := make([]string, l)
		for ix, ident := range root.Names {
			names[ix] = ident.Name
		}

		fmt.Println(indents[depth], "-", strings.Join(names, ", "))
	} else {
		fmt.Println(indents[depth], "<anonymous>")
	}

	context.PrintType(depth+1, root.Type)
}
