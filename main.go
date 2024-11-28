package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Sph3ricalPeter/frbench/external"
	"github.com/Sph3ricalPeter/frbench/external/openai"
	"github.com/Sph3ricalPeter/frbench/internal"
	"github.com/Sph3ricalPeter/frbench/internal/common"
	"github.com/Sph3ricalPeter/frbench/internal/eval"
	"github.com/Sph3ricalPeter/frbench/internal/project"
)

type Mode string

const (
	ModePatchInc Mode = "patch-inc" // patches are applied incrementally for each FR to the codebase
	ModeWriteInc Mode = "write-inc" // writing files instead of patching, also incrementally
	ModeWrite    Mode = "write"     // writing files to a clean codebase for each requirement
)

func main() {
	args := mustParseArgs()

	conns := []external.Connector{
		// openai.NewOpenAIConnector(openai.Gpt4o, ""),
		openai.NewOpenAIConnector(openai.Gpt4oMini, ""),
		// openai.NewOpenAIConnector(openai.O1Mini, ""),
		// google.NewGoogleConnector(google.Gemini15Flash8B, ""),
		// google.NewGoogleConnector(google.Gemini15Pro, ""),
		// anth.NewAnthConnector(anth.Claude3Haiku, ""),
		// anth.NewAnthConnector(anth.Claude35Sonnet, ""),
	}

	benchStats := eval.NewBenchmarkStats()
	outDir := getSnapshotDir(args.evalMode, args.kAttempts, args.temp, benchStats.Timestamp)
	for _, con := range conns {
		modelStats := eval.ModelStats{Projects: map[string]eval.ProjectStats{}}

		for _, projectName := range args.templates {
			projectInfo := project.MustLoadFromYaml(projectName)
			project.MustInitProject(projectName)

			var projectStats eval.ProjectStats
			switch args.mode {
			case ModeWriteInc:
				projectStats = runIncWriteProcedure(args, con, projectInfo, benchStats.Timestamp)
			default:
				panic("invalid mode")
			}

			modelStats.Projects[projectName] = projectStats
		}

		benchStats.Models[con.GetModelName()] = modelStats

		if args.takeSnapshot {
			common.MustWriteJsonFile(benchStats, fmt.Sprintf("%s/stats.json", outDir))
		}
	}

	evalInfo := eval.EvalBenchmark(benchStats, args.evalMode)
	common.MustWriteJsonFile(evalInfo, fmt.Sprintf("%s/eval.json", outDir))
	eval.MustWriteTable(evalInfo, fmt.Sprintf("%s/scores.csv", outDir), func(mps eval.ModelProjectStats) string {
		return fmt.Sprintf("%.1f", mps.Score)
	})
	eval.MustWriteTable(evalInfo, fmt.Sprintf("%s/costs.csv", outDir), func(mps eval.ModelProjectStats) string {
		return fmt.Sprintf("%.5f", mps.Cost)
	})
	eval.MustWriteTable(evalInfo, fmt.Sprintf("%s/resp_times.csv", outDir), func(mps eval.ModelProjectStats) string {
		return fmt.Sprintf("%.5f", mps.Duration.Seconds())
	})
}

// runIncWriteProcedure runs the incremental writing procedure for the given model and project,
//
// each requirement will send a prompt and the expected response is a string containing all changed files
// which will be written to the project's codebase
// still works incrementally in that each requirement will build on the previous one, working with the updated codebase
func runIncWriteProcedure(args Args, con external.Connector, projectInfo project.ProjectInfo, timestamp string) eval.ProjectStats {
	reqCount := len(projectInfo.Project.Requirements)
	projectStats := eval.NewProjectStats(reqCount)

	fmt.Printf("Running write procedure for model %s on project %s ...\n", con.GetModelName(), projectInfo.Project.Name)
	projectFailed := false
	for i := 0; i < reqCount; i++ {
		if !projectFailed {
			reqStats := runReq(args, con, projectInfo, i, timestamp)
			projectStats.Requirements[i] = reqStats

			if !reqStats.Completed {
				fmt.Printf("Requirement %d failed, stopping project ...\n", i+1)
				projectFailed = true
			}
			continue
		}

		projectStats.Requirements[i] = eval.RequirementStats{
			MaxScore: projectInfo.Project.Requirements[i].Score,
		}
	}

	ps := eval.NewProjectSummary(projectStats)

	fmt.Printf("All Done! %s scored: %.2f/%.2f on the %s project!\n", con.GetModelName(), ps.Score, ps.MaxScore, projectInfo.Project.Name)
	fmt.Printf("Total cost for project: $%.5f\n", ps.TotalCost)
	fmt.Printf("Total time for project: %.5fs\n", ps.Duration.Seconds())

	return projectStats
}

// runReq runs the requirement for the given project and model,
// returns error if anything benchmark-breaking happens
func runReq(args Args, con external.Connector, pInfo project.ProjectInfo, i int, timestamp string) eval.RequirementStats {
	req := pInfo.Project.Requirements[i]
	reqStats := eval.RequirementStats{
		MaxScore: req.Score,
	}

	mustFailTest(pInfo, i)

	promptBytes := common.Must(internal.PrepareWritePrompt(pInfo, i))
	imageBytes := common.Must(internal.PrepareImagePrompt(pInfo, i))

	for reIndex := 0; reIndex < args.kAttempts; reIndex++ {
		fmt.Printf("Running requirement %d (re %d/%d)...\n", i+1, reIndex, args.kAttempts-1)

		pOpts := external.NewUserPromptOpts(promptBytes, imageBytes, i+1, args.temp, args.useCache)
		err := runReqAttempt(con, pOpts, &reqStats)

		okMsg := "ok"
		if err != nil {
			okMsg = "fail"
		}

		if args.takeSnapshot {
			project.TakeCodebaseSnapshot(fmt.Sprintf("%s/%s/%s/r%d-a%d-%s", getSnapshotDir(args.evalMode, args.kAttempts, args.temp, timestamp), con.GetModelName(), pInfo.Dir, pOpts.Number, reIndex, okMsg))
		}

		if err == nil {
			reqStats.Completed = true
			break
		}
		fmt.Printf("Error: %s\n", err.Error())

		// wait for input to run re
		if args.interactive {
			fmt.Println("Press ENTER to do next requirement or exit if it's the last one ...")
			_, _ = os.Stdin.Read(make([]byte, 1))
		}
	}

	return reqStats
}

type ReqAttemptResult struct {
	CacheKey *string
}

func runReqAttempt(con external.Connector, pOpts external.SendPromptOpts, reqStats *eval.RequirementStats) error {
	reqStats.Attempts++

	cacheKey := internal.CreateCacheKey(pOpts.Prompt, pOpts.Number)

	fmt.Println("Sending prompt ...")
	result, err := con.SendPrompt(pOpts)

	// prompt failed, retry
	if err != nil {
		return fmt.Errorf("error sending prompt: %w", err)
	}

	if result.UsedCache {
		fmt.Printf("Used cache: %s\n", *result.CacheKey)
	} else {
		cost := external.MustCalcTotalCost(con.GetModelName(), result.Usage, con.GetCost())
		reqStats.Cost += cost
		reqStats.Duration += result.Duration
		fmt.Printf("Req/resp time: %.5fs\n", result.Duration.Seconds())
		fmt.Printf("Used %d input / %d output tokens.\n", result.Usage.InputTokens, result.Usage.OutputTokens)
		if result.Usage.OutputTokens >= external.MaxTokensPerPrompt {
			fmt.Println("Output tokens exceeded the limit! ⚠")
		}
		fmt.Printf("Cost: $%.5f\n", cost)
	}
	fmt.Println("Response OK.")

	// FIXME: testing only
	_ = os.WriteFile("data/resp-content.txt", []byte(result.Content), 0644)

	// parsing failed, retry
	files, err := internal.ParseWriteResponse([]byte(result.Content))
	if err != nil {
		return fmt.Errorf("error parsing write response: %w", err)
	}

	// writing failed, retry (the model could return a bad filepath)
	err = writeAllFiles(files)
	if err != nil {
		return fmt.Errorf("error writing files: %w", err)
	}

	err = runTests()
	// tests after fix failed, retry
	if err != nil {
		fmt.Println("NOK! ❌")
		return fmt.Errorf("test failed after fix")
	}

	fmt.Println("OK! ✅")
	con.CacheResponse(cacheKey, result.RespBytes)

	return nil
}

func writeAllFiles(files []internal.File) error {
	fmt.Printf("Writing %d files ...\n", len(files))
	for _, file := range files {
		fmt.Printf("Writing file %s ...\nContents: %s\n", file.RelPath, file.Content)

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
	return nil
}

func mustFailTest(pInfo project.ProjectInfo, i int) {
	common.CheckErr(copyTestFiles(pInfo.Dir, pInfo.Project.Type, i+1))
	err := runTests()
	if err == nil {
		panic("test passed before patching")
	}
}

// runIncPatchProcedure runs the patching procedure for the given model and project,
//
// each requirement is patched and tested incrementally, with the option to use history to
// send previous prompts and model's responses (similar to a chat conversation)
// func runIncPatchProcedure(args Args, model ModelBenchmark, projectInfo project.ProjectInfo) ProjectStats {
// 	appliedPatches := map[int][]byte{}
// 	invalidPatchCacheKeys := map[int]*string{}

// 	reqCount := len(projectInfo.Project.Requirements)
// 	for i := 1; i < reqCount+1; i++ {
// 		fmt.Printf("Running requirement #%d: %s on model %s ...\n", i, projectInfo.Project.Requirements[i-1].Name, model.con.GetModelName())

// 		// move test file before prompt is created with codebase inside
// 		err := copyTestFiles(projectInfo.Dir, projectInfo.Project.Type, i)
// 		if err != nil {
// 			fmt.Printf("error copying test file: %s\n", err)
// 			break
// 		}

// 		// tests should fail before patching
// 		err = runTests()
// 		if err == nil {
// 			panic("test passed before patching")
// 		}

// 		promptBytes, err := internal.PreparePatchPrompt(projectInfo, i-1)
// 		if err != nil {
// 			fmt.Printf("error reading test file: %s\n", err.Error())
// 			break
// 		}
// 		fmt.Println("Sending prompt ...")
// 		result, err := model.con.SendPrompt(external.SendPromptOpts{
// 			Number:   i,
// 			Role:     external.RoleUser,
// 			Prompt:   promptBytes,
// 			UseCache: args.useCache,
// 			// FIXME: we are sending the whole codebase with each prompt and not using history
// 			UseHistory: false,
// 		})
// 		if err != nil {
// 			fmt.Printf("error sending prompt: %s\n", err.Error())
// 			break
// 		}
// 		if result.UsedCache {
// 			fmt.Printf("Used cache: %s\n", *result.CacheKey)
// 		} else {
// 			fmt.Printf("Used %d input / %d output tokens.\n", result.Usage.InputTokens, result.Usage.OutputTokens)
// 		}
// 		fmt.Println("OK.")

// 		updatedPatch, err := internal.Patch([]byte(result.Content), i)
// 		if err != nil {
// 			fmt.Printf("error doing patch: %s\n", err.Error())
// 			invalidPatchCacheKeys[i] = result.CacheKey
// 			break
// 		}
// 		cleanupWeirdFiles()
// 		appliedPatches[i] = updatedPatch

// 		err = runTests()
// 		if err != nil {
// 			fmt.Println("Patch BAD! ❌")
// 			invalidPatchCacheKeys[i] = result.CacheKey
// 			break
// 		} else {
// 			fmt.Println("Patch OK! ✅")
// 			model.stats[projectInfo.Project.Name]++
// 		}

// 		// wait for input to revert patches
// 		fmt.Println("Press ENTER to do next patch or revert if it's the last one ...")
// 		_, _ = os.Stdin.Read(make([]byte, 1))
// 	}

// 	fmt.Printf("All Done! %s scored: %d/%d\n", model.con.GetModelName(), model.stats[projectInfo.Project.Name], reqCount)

// 	// prompt is the key for the cache
// 	for i, cacheKey := range invalidPatchCacheKeys {
// 		if cacheKey == nil {
// 			continue
// 		}
// 		fmt.Printf("Removing cache for invalid patch #%d ...\n", i)
// 		common.CheckErr(model.con.InvalidateCachedPrompt(*cacheKey))
// 	}

// 	return ProjectStats{}
// }

// func cleanupWeirdFiles() {
// 	fmt.Println("Removing weird files ...")
// 	err := common.RunBashCommand("rm -f app/*.go.orig")
// 	if err != nil {
// 		fmt.Printf("error removing weird files: %s\n", err.Error())
// 	}
// }

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

type Args struct {
	newProject   string
	templates    []string
	mode         Mode
	kAttempts    int
	evalMode     eval.EvalMode
	temp         float64
	useCache     bool
	interactive  bool
	takeSnapshot bool
}

func NewArgs(new string, templates []string, mode Mode, kAttempts int, evalMode eval.EvalMode, temp float64, useCache, interactive, takeSnapshot bool) Args {
	return Args{
		newProject:   new,
		templates:    templates,
		mode:         mode,
		kAttempts:    kAttempts,
		evalMode:     evalMode,
		temp:         temp,
		useCache:     useCache,
		interactive:  interactive,
		takeSnapshot: takeSnapshot,
	}
}

func mustParseArgs() Args {
	fsCreate := flag.NewFlagSet("create", flag.ExitOnError)
	newArg := fsCreate.String("name", "simple-todo", "project name to use")

	if len(os.Args) < 2 {
		panic("Bad usage. Use -h for help.")
	}

	if os.Args[1] == "create" {
		fsCreate.Parse(os.Args[2:])
		newProject := *newArg
		fmt.Println("new project:", newProject)
		if newProject != "" {
			project.MustInitTemplate(newProject, project.ProjectTypeSingle)
		}
		os.Exit(0)
	}

	tArg := flag.String("t", "functions", "project template to use (functions | simple-todo)")
	mArg := flag.String("M", "write-inc", "mode to run (patch-inc | write-inc | write)")
	cArg := flag.Bool("c", false, "use cache")
	iArg := flag.Bool("i", false, "interactive mode")
	sArg := flag.Bool("s", false, "take snapshot")
	eArg := flag.String("e", "weighted-pass-k", "evaluation mode to use (weighted-score-k | score-k)")
	kArg := flag.Int("k", 3, "number of attempts to make for each requirement")
	tempArg := flag.Float64("T", 0.7, "temperature to use for the model")
	// modelsArg := flag.String("m", "gpt4o-mini", "models to use (gpt4o-mini | gpt4o | o1-mini | gemini15-flash-8b | gemini15-pro | claude3-haiku | claude35-sonnet)")
	flag.Parse()
	args := NewArgs(*newArg, parseProjectsArg(*tArg), Mode(*mArg), *kArg, eval.EvalMode(*eArg), *tempArg, *cArg, *iArg, *sArg)
	fmt.Printf("Running with args: %+v\n", args)

	return args
}

// parseProjectsArg parses comma separated project names
func parseProjectsArg(arg string) []string {
	return strings.Split(arg, ",")
}

func getSnapshotDir(evalMode eval.EvalMode, kAttempts int, temp float64, timestamp string) string {
	return fmt.Sprintf("out/%s_k%d_T%02.0f_%s", evalMode, kAttempts, temp*10, timestamp)
}
