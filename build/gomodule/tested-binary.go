package gomodule

import (
	"fmt"
	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
	"path"
)

var (
	// Package context used to define Ninja build rules.
	pctx = blueprint.NewPackageContext("github.com/EugeniaKol/design-practice-1/build/gomodule")

	// Ninja rule to execute go build.
	goBuild = pctx.StaticRule("binaryBuild", blueprint.RuleParams{
		Command:     "cd $workDir && go build -o $outputPath $pkg",
		Description: "build go command $pkg",
	}, "workDir", "outputPath", "pkg")

	// Ninja rule to execute go mod vendor.
	goVendor = pctx.StaticRule("vendor", blueprint.RuleParams{
		Command:     "cd $workDir && go mod vendor",
		Description: "vendor dependencies of $name",
	}, "workDir", "name")

	// Ninja rule to execute go test.
	goTest = pctx.StaticRule("test", blueprint.RuleParams{
		Command:     "cd $workDir && go test -v $pkg > $outputPath",
		Description: "test $pkg",
	}, "workDir", "pkg", "outputPath")
)

type testedBinaryModule struct {
	blueprint.SimpleName
	properties struct {
		Pkg     string
		TestPkg string

		// List of source files - different source files for test and build
		Srcs     []string
		TestSrcs []string
		// Exclude patterns.
		SrcsExclude     []string
		TestSrcsExclude []string
		// If to call vendor command.
		VendorFirst bool

		// Example of how to specify dependencies.
		Deps []string
	}
}

func (tb *testedBinaryModule) DynamicDependencies(blueprint.DynamicDependerModuleContext) []string {
	return tb.properties.Deps
}

func (tb *testedBinaryModule) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	config := bood.ExtractConfig(ctx)
	config.Debug.Printf("Adding build actions and testing for go binary module '%s'", name)

	binOutputPath := path.Join(config.BaseOutputDir, "bin", name)

	var inputs []string
	inputErors := false
	for _, src := range tb.properties.Srcs {
		//avoiding building if only tests are changed
		if matches, err := ctx.GlobWithDeps(src, append(tb.properties.SrcsExclude, tb.properties.TestSrcs...)); err == nil {
			inputs = append(inputs, matches...)
		} else {
			ctx.PropertyErrorf("srcs", "Cannot resolve files that match pattern %s", src)
			inputErors = true
		}
	}
	if inputErors {
		return
	}

	testOutputPath := path.Join(config.BaseOutputDir, "test", name)
	//getting sources for testing
	var testInputs []string

	for _, src := range tb.properties.TestSrcs {
		if matches, err := ctx.GlobWithDeps(src, tb.properties.TestSrcsExclude); err == nil {
			testInputs = append(testInputs, matches...)
		} else {
			ctx.PropertyErrorf("testSrcs", "Cannot resolve files that match pattern %s", src)
			inputErors = true
		}
	}

	if inputErors {
		return
	}
	//if build files are changed, tests are run too
	testInputs = append(testInputs, inputs...)

	if tb.properties.VendorFirst {
		vendorDirPath := path.Join(ctx.ModuleDir(), "vendor")
		ctx.Build(pctx, blueprint.BuildParams{
			Description: fmt.Sprintf("Vendor dependencies of %s", name),
			Rule:        goVendor,
			Outputs:     []string{vendorDirPath},
			Implicits:   []string{path.Join(ctx.ModuleDir(), "go.mod")},
			Optional:    true,
			Args: map[string]string{
				"workDir": ctx.ModuleDir(),
				"name":    name,
			},
		})
		inputs = append(inputs, vendorDirPath)
	}

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Build %s as Go binary", name),
		Rule:        goBuild,
		Outputs:     []string{binOutputPath},
		Implicits:   inputs,
		Args: map[string]string{
			"outputPath": binOutputPath,
			"workDir":    ctx.ModuleDir(),
			"pkg":        tb.properties.Pkg,
		},
	})

	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Run tests of %s", name),
		Rule:        goTest,
		Outputs:     []string{testOutputPath},
		Implicits:   testInputs,
		Args: map[string]string{
			"outputPath": testOutputPath,
			"workDir":    ctx.ModuleDir(),
			"pkg":        tb.properties.TestPkg,
		},
	})
}

func TestedBinFactory() (blueprint.Module, []interface{}) {
	mType := &testedBinaryModule{}
	return mType, []interface{}{&mType.SimpleName.Properties, &mType.properties}
}
