package cmd

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

func load() {
	var yolo v1.Pod
	pretty.Println("yolo", yolo)

	var conf loader.Config

	// Add "runtime" to the set of packages to be loaded.
	conf.Import("k8s.io/api/core/v1")

	// Finally, load all the packages specified by the configuration.
	program, err := conf.Load()

	if err != nil {
		glog.Fatal(err)
	}

	//pkg := program.InitialPackages()[0]
	//PrintImportSpecs(program)

	//context := ContextForType(program, "k8s.io/api/core/v1", "Pod")
	contexts := ContextsForPackage(program, program.InitialPackages()[0])

	for _, context := range contexts {
		context.Print(0)
	}

	fmt.Println("yes, hello")
}

type Context struct {
	Program  *loader.Program
	Package  *loader.PackageInfo
	File     *ast.File
	TypeSpec *ast.TypeSpec
}

func importMatchesPackage(imprt *ast.ImportSpec, pkg *types.Package) bool {
	quoted := imprt.Path.Value
	stripped := quoted[1 : len(quoted)-1]
	return pathMatchesPackage(stripped, pkg)
}

func pathMatchesPackage(pkgPath string, pkg *types.Package) bool {
	if pkgPath == pkg.Path() {
		return true
	}

	if strings.HasSuffix(pkg.Path(), "vendor/"+pkgPath) {
		return true
	}

	return false
}

func ContextForImportedType(program *loader.Program, imprt *ast.ImportSpec, typeName string) *Context {
	quoted := imprt.Path.Value
	stripped := quoted[1 : len(quoted)-1]
	return ContextForType(program, stripped, typeName)
}

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

func (context *Context) RefocusedWithSelector(selector *ast.SelectorExpr) *Context {
	var pkgName string
	switch expr := selector.X.(type) {
	case *ast.Ident:
		pkgName = expr.Name
	default:
		glog.Fatal(pretty.Sprint(selector))
	}

	typeName := selector.Sel.Name

	return context.RefocusedWithPkgAndTypeNames(pkgName, typeName)
}

func (context *Context) GetImportedPackage(imprt *ast.ImportSpec) *loader.PackageInfo {
	for _, pkg := range context.Program.AllPackages {
		if importMatchesPackage(imprt, pkg.Pkg) {
			return pkg
		}
	}

	return nil
}

func (context *Context) RefocusedWithPkgAndTypeNames(pkgName string, typeName string) *Context {
	for _, imprt := range context.File.Imports {
		if imprt.Name != nil {
			// If the local name matches, look up the type here.
			if imprt.Name.Name == pkgName {
				return ContextForImportedType(context.Program, imprt, typeName)
			}
		} else {
			// If the default name matches, look up the type here.
			pkg := context.GetImportedPackage(imprt)
			if pkg.Pkg.Name() == pkgName {
				return ContextForPackageAndType(context.Program, pkg, typeName)
			}
		}
	}

	return nil
}

var maxDepth = 40
var indents = make([]string, maxDepth)

func init() {
	for index := range indents {
		indents[index] = strings.Repeat("  ", index)
	}
}

func (context *Context) Print(depth int) {
	if context == nil {
		glog.Fatal()
	}

	fmt.Println(indents[depth], context.TypeSpec.Name)
	context.PrintType(depth+1, context.TypeSpec.Type)
}

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
	}
}

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
