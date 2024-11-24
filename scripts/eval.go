package main

import (
	"fmt"

	"github.com/Sph3ricalPeter/frbench/internal/common"
	"github.com/Sph3ricalPeter/frbench/internal/eval"
)

const (
	BenchDir = "pass-k_k10_T02_2024-11-23_20-45-41"
)

func main() {
	outDir := fmt.Sprintf("out/%s", BenchDir)

	benchStats := common.MustReadJsonFileInto[eval.BenchmarkStats](fmt.Sprintf("%s/stats.json", outDir))

	evalInfo := eval.EvalBenchmark(benchStats, eval.PassK)

	common.MustWriteJsonFile(evalInfo, fmt.Sprintf("out/%s/eval-1.json", BenchDir))
	eval.MustWriteTable(evalInfo, fmt.Sprintf("%s/scores-1.csv", outDir), func(mps eval.ModelProjectStats) string {
		return fmt.Sprintf("%.1f", mps.Score)
	})
	eval.MustWriteTable(evalInfo, fmt.Sprintf("%s/costs-1.csv", outDir), func(mps eval.ModelProjectStats) string {
		return fmt.Sprintf("%.5f", mps.Cost)
	})
	eval.MustWriteTable(evalInfo, fmt.Sprintf("%s/resp_times-1.csv", outDir), func(mps eval.ModelProjectStats) string {
		return fmt.Sprintf("%.5f", mps.Duration.Seconds())
	})
}
