package gomodule

import (
	"fmt"
	"github.com/google/blueprint"
	"github.com/roman-mazur/bood"
	"path"
)

var (
	goArchive = pctx.StaticRule("goArchive", blueprint.RuleParams{
		Command:     "cd $workDir && zip -r -j $outputPath $input",
		Description: "archive results of $binary",
	}, "workDir", "outputPath", "input", "binary")
)

type archiveModule struct {
	blueprint.SimpleName
	properties struct {
		//name of binary to archive
		Binary string
	}
}

func (am *archiveModule) DynamicDependencies(blueprint.DynamicDependerModuleContext) []string {
	return []string{am.properties.Binary}
}

func (am *archiveModule) GenerateBuildActions(ctx blueprint.ModuleContext) {
	name := ctx.ModuleName()
	config := bood.ExtractConfig(ctx)
	config.Debug.Printf("Adding archive actions for go binary module '%s'", name)

	binName := am.properties.Binary
	baseDir := config.BaseOutputDir

	//getting results of other module using its name
	binModule, _ := ctx.GetDirectDep(binName)
	if binModule == nil {
		println("cant find ", binName, " dependency")
	}

	inputs := binModule.(*testedBinaryModule).GetBinPath(baseDir)

	zipOutputPath := path.Join(baseDir, "archive", name)

	//adding rule for archiving
	ctx.Build(pctx, blueprint.BuildParams{
		Description: fmt.Sprintf("Archive results of %s module", binName),
		Rule:        goArchive,
		Outputs:     []string{zipOutputPath},
		Implicits:   []string{inputs},
		Args: map[string]string{
			"outputPath": zipOutputPath,
			"workDir":    ctx.ModuleDir(),
			"input":      inputs,
			"binary":     binName,
		},
	})
}

func ArchiverFactory() (blueprint.Module, []interface{}) {
	mType := &archiveModule{}
	return mType, []interface{}{&mType.SimpleName.Properties, &mType.properties}
}
