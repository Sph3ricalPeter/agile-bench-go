package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/Sph3ricalPeter/frbench/external"
	"github.com/Sph3ricalPeter/frbench/external/google"
	"github.com/Sph3ricalPeter/frbench/internal"
	"github.com/Sph3ricalPeter/frbench/internal/common"
	"github.com/Sph3ricalPeter/frbench/internal/project"
)

type ProjectTemplate string

const (
	ProjectFunctions  ProjectTemplate = "functions"
	ProjectSimpleTodo ProjectTemplate = "simple-todo"
)

type Mode string

const (
	ModePatchInc Mode = "patch-inc" // patches are applied incrementally for each FR to the codebase
	ModeWriteInc Mode = "write-inc" // writing files instead of patching, also incrementally
	ModeWrite    Mode = "write"     // writing files to a clean codebase for each requirement
)

type Args struct {
	template ProjectTemplate
	mode     Mode
	useCache bool
}

func NewArgs(template ProjectTemplate, mode Mode, useCache bool) Args {
	return Args{
		template: template,
		mode:     mode,
		useCache: useCache,
	}
}

type ModelBenchmark struct {
	con   external.Connector
	stats map[string]int // score = 1 / triesTillPass
}

func NewModelBenchmark(con external.Connector) ModelBenchmark {
	return ModelBenchmark{
		con:   con,
		stats: make(map[string]int),
	}
}

func main() {
	// read cache flag from args
	tArg := flag.String("t", "functions", "project template to use (functions | simple-todo)")
	mArg := flag.String("m", "write-inc", "mode to run (patch-inc | write-inc | write)")
	cArg := flag.Bool("c", false, "use cache")
	flag.Parse()
	args := NewArgs(ProjectTemplate(*tArg), Mode(*mArg), *cArg)
	fmt.Printf("Running with args: %+v\n", args)

	models := []ModelBenchmark{
		NewModelBenchmark(google.NewGoogleConnector(google.Gemini15Flash8B, "")),
		// NewModelBenchmark(anth.NewAnthConnector(anth.Claude3Haiku, "")),
	}

	// 1. copy the initial codebase for the project
	project.MustInitProject(string(args.template))

	// 2. read project.yml and load the project info
	projectInfo := project.MustLoadFromYaml(string(args.template))

	for _, model := range models {
		switch args.mode {
		case ModePatchInc:
			runIncPatchProcedure(args, model, projectInfo)
		case ModeWriteInc:
			runIncWriteProcedure(args, model, projectInfo)
		default:
			panic("invalid mode")
		}
	}
}

// runIncWriteProcedure runs the incremental writing procedure for the given model and project,
//
// each requirement will send a prompt and the expected response is a string containing all changed files
// which will be written to the project's codebase
// still works incrementally in that each requirement will build on the previous one, working with the updated codebase
func runIncWriteProcedure(args Args, model ModelBenchmark, projectInfo project.ProjectInfo) {
	reqCount := len(projectInfo.Project.Requirements)

	invalidRespCacheKeys := map[int]*string{}
	fmt.Printf("Running write procedure for model %s on project %s ...\n", model.con.GetModelName(), projectInfo.Project.Name)
	for i := 0; i < reqCount; i++ {
		req := projectInfo.Project.Requirements[i]
		fmt.Printf("Running requirement #%d: %s ...\n", i, req.Name)

		err := copyTestFiles(projectInfo.Dir, projectInfo.Project.Type, i+1)
		if err != nil {
			fmt.Printf("error copying test file: %s\n", err)
			break
		}

		err = runTests()
		if err == nil {
			panic("test passed before patching")
		}

		promptBytes, err := internal.PrepareWritePrompt(projectInfo, i)
		if err != nil {
			fmt.Printf("error preparing prompt: %s\n", err.Error())
			break
		}
		fmt.Println("Sending prompt ...")
		result, err := model.con.SendPrompt(external.SendPromptOpts{
			Number:     i + 1,
			Role:       external.RoleUser,
			Prompt:     promptBytes,
			UseCache:   args.useCache,
			UseHistory: false,
		})
		if err != nil {
			fmt.Printf("error sending prompt: %s\n", err.Error())
			break
		}
		if result.UsedCache {
			fmt.Printf("Used cache: %s\n", *result.CacheKey)
		} else {
			fmt.Printf("Used %d input / %d output tokens.\n", result.Usage.InputTokens, result.Usage.OutputTokens)
		}
		fmt.Println("OK.")

		files, err := internal.ParseWriteResponse([]byte(result.Content))
		if err != nil {
			fmt.Printf("error parsing write response: %s\n", err.Error())
			break
		}

		for _, file := range files {
			path := filepath.Join("app/", file.RelPath)
			err := os.MkdirAll(filepath.Dir(path), os.ModePerm)
			if err != nil {
				fmt.Printf("error creating app directory: %s\n", err.Error())
				break
			}
			err = os.WriteFile(fmt.Sprintf("app/%s", file.RelPath), file.Content, 0644)
			if err != nil {
				fmt.Printf("error writing file: %s\n", err.Error())
			}
		}

		err = runTests()
		if err != nil {
			fmt.Println("NOK! ❌")
			invalidRespCacheKeys[i] = result.CacheKey
			break
		} else {
			fmt.Println("OK! ✅")
			model.stats[projectInfo.Project.Name]++
		}

		// wait for input to revert patches
		fmt.Println("Press ENTER to do next requirement or exit if it's the last one ...")
		_, _ = os.Stdin.Read(make([]byte, 1))
	}

	fmt.Printf("All Done! %s scored: %d/%d on the %s project!\n", model.con.GetModelName(), model.stats[projectInfo.Project.Name], reqCount, projectInfo.Project.Name)

	// prompt is the key for the cache
	for i, cacheKey := range invalidRespCacheKeys {
		if cacheKey == nil {
			continue
		}
		fmt.Printf("Removing cache for invalid resp. #%d ...\n", i)
		common.CheckErr(model.con.InvalidateCachedPrompt(*cacheKey))
	}
}

// runIncPatchProcedure runs the patching procedure for the given model and project,
//
// each requirement is patched and tested incrementally, with the option to use history to
// send previous prompts and model's responses (similar to a chat conversation)
func runIncPatchProcedure(args Args, model ModelBenchmark, projectInfo project.ProjectInfo) {
	appliedPatches := map[int][]byte{}
	invalidPatchCacheKeys := map[int]*string{}

	reqCount := len(projectInfo.Project.Requirements)
	for i := 1; i < reqCount+1; i++ {
		fmt.Printf("Running requirement #%d: %s on model %s ...\n", i, projectInfo.Project.Requirements[i-1].Name, model.con.GetModelName())

		// move test file before prompt is created with codebase inside
		err := copyTestFiles(projectInfo.Dir, projectInfo.Project.Type, i)
		if err != nil {
			fmt.Printf("error copying test file: %s\n", err)
			break
		}

		// tests should fail before patching
		err = runTests()
		if err == nil {
			panic("test passed before patching")
		}

		promptBytes, err := internal.PreparePatchPrompt(projectInfo, i-1)
		if err != nil {
			fmt.Printf("error reading test file: %s\n", err.Error())
			break
		}
		fmt.Println("Sending prompt ...")
		result, err := model.con.SendPrompt(external.SendPromptOpts{
			Number:   i,
			Role:     external.RoleUser,
			Prompt:   promptBytes,
			UseCache: args.useCache,
			// FIXME: we are sending the whole codebase with each prompt and not using history
			UseHistory: false,
		})
		if err != nil {
			fmt.Printf("error sending prompt: %s\n", err.Error())
			break
		}
		if result.UsedCache {
			fmt.Printf("Used cache: %s\n", *result.CacheKey)
		} else {
			fmt.Printf("Used %d input / %d output tokens.\n", result.Usage.InputTokens, result.Usage.OutputTokens)
		}
		fmt.Println("OK.")

		updatedPatch, err := internal.Patch([]byte(result.Content), i)
		if err != nil {
			fmt.Printf("error doing patch: %s\n", err.Error())
			invalidPatchCacheKeys[i] = result.CacheKey
			break
		}
		cleanupWeirdFiles()
		appliedPatches[i] = updatedPatch

		err = runTests()
		if err != nil {
			fmt.Println("Patch BAD! ❌")
			invalidPatchCacheKeys[i] = result.CacheKey
			break
		} else {
			fmt.Println("Patch OK! ✅")
			model.stats[projectInfo.Project.Name]++
		}

		// wait for input to revert patches
		fmt.Println("Press ENTER to do next patch or revert if it's the last one ...")
		_, _ = os.Stdin.Read(make([]byte, 1))
	}

	fmt.Printf("All Done! %s scored: %d/%d\n", model.con.GetModelName(), model.stats[projectInfo.Project.Name], reqCount)

	// prompt is the key for the cache
	for i, cacheKey := range invalidPatchCacheKeys {
		if cacheKey == nil {
			continue
		}
		fmt.Printf("Removing cache for invalid patch #%d ...\n", i)
		common.CheckErr(model.con.InvalidateCachedPrompt(*cacheKey))
	}
}

func cleanupWeirdFiles() {
	fmt.Println("Removing weird files ...")
	err := common.RunBashCommand("rm -f app/*.go.orig")
	if err != nil {
		fmt.Printf("error removing weird files: %s\n", err.Error())
	}
}

// copyTestFile copies the test file for the given requirement number to the app/ directory
//
// depending on the project type, the test file is copied from the reference directory
// for single projects, the test file is copied directly
// for checkpoint projects, all test files for the given checkpoint are copied and the existing test files are removed
func copyTestFiles(templateDir string, projectType project.ProjectType, reqN int) error {
	switch projectType {
	case project.ProjectTypeSingle:
		// single test for a given requirement number is copied
		return common.RunCommand(fmt.Sprintf("cp templates/%s/reference/%d_test.go app/", templateDir, reqN))
	case project.ProjectTypeCheckpoints:
		// all existing test files are removed (in case of breaking changes)
		err := common.RunBashCommand("rm -f app/*_test.go")
		if err != nil {
			return fmt.Errorf("error removing test files: %w", err)
		}
		// all test files for the given checkpoint are copied
		return common.RunBashCommand(fmt.Sprintf("cp templates/%s/reference/%d/*_test.go app/", templateDir, reqN))
	}
	return fmt.Errorf("invalid project type")
}

func runTests() error {
	return common.RunCommand("go test ./app/")
}
