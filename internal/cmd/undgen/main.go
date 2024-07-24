package main

import (
	"flag"
	"fmt"
	"go/printer"
	"os"
	"path"

	"github.com/ngicks/und/internal/undgen"
	"golang.org/x/tools/go/packages"
)

var (
	inputPkg    = flag.String("i", "", "input package name. must be rooted or relative path. it's ok to be ./...")
	outFilename = flag.String("o", "", "base name for output file.")
)

func main() {
	flag.Parse()

	cfg := &packages.Config{
		Mode: packages.NeedFiles | packages.NeedSyntax | packages.NeedImports |
			packages.NeedDeps | packages.NeedExportFile | packages.NeedTypes |
			packages.NeedSyntax | packages.NeedTypesInfo | packages.NeedModule |
			packages.NeedName, // almost all load bits. We'll reduce option as many as possible.
	}
	pkgs, err := packages.Load(cfg, *inputPkg)
	if err != nil {
		panic(err)
	}

	fmt.Printf("pkgs: %#v\n", pkgs)

	targets, err := undgen.TargetTypes(pkgs)
	if err != nil {
		panic(err)
	}
	fmt.Printf("targets: %#v\n", targets)

	gen, err := undgen.GeneratePlainType(pkgs)
	if err != nil {
		panic(err)
	}

	for k, v := range gen.Pkg {
		pkgName := v.PkgName
		if pkgName == "" {
			pkgName = path.Base(k)
		}
		fmt.Printf("package %s\n\n", pkgName)
		fmt.Printf("import (\n")
		for k, v := range v.Imports {
			fmt.Printf("\t")
			if v != "" {
				fmt.Printf("%s ", v)
			}
			fmt.Printf("%q\n", k)
		}
		fmt.Printf(")\n\n")

		for _, ty := range v.Generated {
			fmt.Printf("//undgen:generated\n")
			printer.Fprint(os.Stdout, ty.Fset, ty.Decl)
			fmt.Printf("\n\n")
		}
	}
}
