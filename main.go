package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/Sph3ricalPeter/frbench/external"
	"github.com/Sph3ricalPeter/frbench/external/google"
	"github.com/Sph3ricalPeter/frbench/internal"
	"github.com/Sph3ricalPeter/frbench/internal/common"
	"github.com/Sph3ricalPeter/frbench/internal/project"
)

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
	useCache := flag.Bool("c", false, "use cache")
	flag.Parse()

	models := []ModelBenchmark{
		NewModelBenchmark(google.NewGoogleConnector(google.Gemini15Flash8B, "")),
		// NewModelBenchmark(anth.NewAnthConnector(anth.Claude3Haiku, "")),
	}

	// 1. copy the initial codebase for the project
	projectName := "functions"
	project.MustInitProject(projectName)

	// 2. read project.yml and load the project info
	projectInfo := project.MustLoadFromYaml(projectName)

	appliedPatches := map[int][]byte{}
	invalidPatchCacheKeys := map[int]*string{}
	reqCount := len(projectInfo.Project.Requirements)
	for _, model := range models {
		for i := 1; i < reqCount+1; i++ {
			fmt.Printf("Running requirement #%d: %s on model %s ...\n", i, projectInfo.Project.Requirements[i-1].Name, model.con.GetModelName())

			// move test file before prompt is created with codebase inside
			err := copyTestFile(i)
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
				UseCache: *useCache,
				// FIXME: we are sending the whole codebase with each prompt and not using history
				UseHistory: false,
			})
			if err != nil {
				fmt.Printf("error sending prompt: %s\n", err.Error())
				break
			}
			fmt.Println("OK.")
			if result.CacheKey == nil {
				fmt.Printf("Used %d input / %d output tokens.\n", result.Usage.InputTokens, result.Usage.OutputTokens)
			}

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
}

func cleanupWeirdFiles() {
	fmt.Println("Removing weird files ...")
	err := common.RunBashCommand("rm -f app/*.go.orig")
	if err != nil {
		fmt.Printf("error removing weird files: %s\n", err.Error())
	}
}

func copyTestFile(i int) error {
	return common.RunCommand(fmt.Sprintf("cp templates/functions/reference/%d_test.go app/", i))
}

func runTests() error {
	return common.RunCommand("go test -v ./app/")
}
