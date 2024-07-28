package main

import (
	"context"
	"flag"
	"fmt"

	"github.com/ngicks/und/internal/undgen"
	"golang.org/x/tools/go/packages"
)

var (
	inputPkg          = flag.String("i", "", "input package name. must be rooted or relative path. it's ok to be ./...")
	outFilenameSuffix = flag.String("o", "", "base name for output file.")
)

const (
	// this is merely a place holder.
	generatorPkgName = "github.com/ngicks/und/internal/cmd/undgen"
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

	p := undgen.Printer{
		GeneratorPkgName: generatorPkgName,
		FileSuffix:       *outFilenameSuffix,
	}

	err = p.Print(context.Background(), gen)
	if err != nil {
		panic(err)
	}
}
