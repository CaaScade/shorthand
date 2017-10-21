package cmd

import (
	"fmt"
	"go/ast"
	"go/token"
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

	/*
		// Use the command-line arguments to specify
		// a set of initial packages to load from source.
		// See FromArgsUsage for help.
		rest, err := conf.FromArgs(os.Args[1:], wantTests)
	*/

	// Add "runtime" to the set of packages to be loaded.
	conf.Import("k8s.io/api/core/v1")

	// Finally, load all the packages specified by the configuration.
	program, err := conf.Load()

	if err != nil {
		glog.Fatal(err)
	}

	//pkg := program.InitialPackages()[0]

	context := ContextForType(program, "Pod")
	fmt.Println(context)
	context.PrintTypeSpec(0, context.TypeSpec)

	fmt.Println("yes, hello")
}

type PrintContext struct {
	Program  *loader.Program
	File     *ast.File
	TypeSpec *ast.TypeSpec
}

func ContextForType(program *loader.Program, typeName string) *PrintContext {
	for _, pkg := range program.InitialPackages() {
		for _, file := range pkg.Files {
			for _, decl := range file.Decls {
				switch decl := decl.(type) {
				case *ast.GenDecl:
					if decl.Tok == token.TYPE {
						for _, spec := range decl.Specs {
							switch spec := spec.(type) {
							case *ast.TypeSpec:
								if spec.Name.Name == typeName {
									return &PrintContext{Program: program, File: file, TypeSpec: spec}
								}
							}
						}
					}
				}
			}
		}
	}

	return nil
}

func GetTypeSpec(pkg *loader.PackageInfo, typeName string) *ast.TypeSpec {
	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				if decl.Tok == token.TYPE {
					for _, spec := range decl.Specs {
						switch spec := spec.(type) {
						case *ast.TypeSpec:
							return spec
						}
					}
				}
			}
		}
	}

	return nil
}

func (context *PrintContext) GetImportedPackage(imprt *ast.ImportSpec) *loader.PackageInfo {
	for _, pkg := range context.Program.AllPackages {
		if imprt.Path.Value != pkg.Pkg.Path() {
			return pkg
		}
	}

	return nil
}

func (context *PrintContext) GetImportedTypeSpec(pkgName string, typeName string) *ast.TypeSpec {
	for _, imprt := range context.File.Imports {
		if imprt.Name != nil {
			fmt.Println(imprt.Name.Name)
			fmt.Println(imprt.Path.Value)
			// If the local name matches, look up the type here.
			if imprt.Name.Name == pkgName {
				pkg := context.GetImportedPackage(imprt)
				return GetTypeSpec(pkg, typeName)
			}
		} else {
			// If the default name matches, look up the type here.
			pkg := context.GetImportedPackage(imprt)
			fmt.Println(pkg.Pkg.Name())
			fmt.Println(imprt.Path.Value)
			if pkg.Pkg.Name() == pkgName {
				return GetTypeSpec(pkg, typeName)
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

func (context *PrintContext) GetSelectorTypeSpec(selector *ast.SelectorExpr) *ast.TypeSpec {
	var pkgName string
	switch expr := selector.X.(type) {
	case *ast.Ident:
		pkgName = expr.Name
	default:
		glog.Fatal(pretty.Sprint(selector))
	}

	typeName := selector.Sel.Name

	return context.GetImportedTypeSpec(pkgName, typeName)
}

func (context *PrintContext) PrintTypeSpec(depth int, root *ast.TypeSpec) {
	fmt.Println(indents[depth], root.Name)
	context.PrintType(depth+1, root.Type)
}

func (context *PrintContext) PrintType(depth int, root ast.Expr) {
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

func (context *PrintContext) PrintSelector(depth int, root *ast.SelectorExpr) {
	var pkgName string
	switch expr := root.X.(type) {
	case *ast.Ident:
		pkgName = expr.Name
	default:
		glog.Fatal(pretty.Sprint(root))
	}

	typeName := root.Sel.Name

	fmt.Printf("%s%s.%s\n", indents[depth], pkgName, typeName)
	typeSpec := context.GetImportedTypeSpec(pkgName, typeName)
	if typeSpec != nil {
		context.PrintTypeSpec(depth, typeSpec)
	}
}

func (context *PrintContext) PrintTypeIdent(depth int, root *ast.Ident) {
	obj := root.Obj
	if obj != nil {
		switch obj.Kind {
		case ast.Typ:
			switch decl := obj.Decl.(type) {
			case *ast.TypeSpec:
				context.PrintTypeSpec(depth, decl)
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

func (context *PrintContext) PrintField(depth int, root *ast.Field) {
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
