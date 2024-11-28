package eval

import (
	"encoding/csv"
	"maps"
	"os"
	"sort"
	"time"
)

type BenchmarkStats struct {
	Timestamp string                `json:"timestamp"`
	Models    map[string]ModelStats `json:"models"`
}

func NewBenchmarkStats() BenchmarkStats {
	return BenchmarkStats{
		Timestamp: time.Now().Format("2006-01-02_15-04-05"),
		Models:    map[string]ModelStats{},
	}
}

type ModelStats struct {
	Projects map[string]ProjectStats `json:"projects"`
}

type ProjectStats struct {
	Requirements []RequirementStats `json:"requirements"`
}

func NewProjectStats(reqCount int) ProjectStats {
	return ProjectStats{
		Requirements: make([]RequirementStats, reqCount),
	}
}

type RequirementStats struct {
	Cost      float64       `json:"cost"`
	Completed bool          `json:"completed"`
	MaxScore  int           `json:"max_score"`
	Attempts  int           `json:"attempts"`
	Duration  time.Duration `json:"duration"`
}

type ProjectSummary struct {
	Score     float64       `json:"score"`
	MaxScore  float64       `json:"max_score"`
	TotalCost float64       `json:"cost"`
	Duration  time.Duration `json:"duration"`
}

func NewProjectSummary(ps ProjectStats) ProjectSummary {
	maxScore := 0.0
	score := 0.0
	cost := 0.0
	d := time.Duration(0)
	for _, reqStats := range ps.Requirements {
		maxScore += float64(reqStats.MaxScore)
		if reqStats.Completed {
			score += float64(reqStats.MaxScore) / float64(reqStats.Attempts)
		}
		cost += reqStats.Cost
		d += reqStats.Duration
	}
	return ProjectSummary{
		Score:     score,
		MaxScore:  maxScore,
		TotalCost: cost,
		Duration:  d,
	}
}

type EvalMode string

const (
	ScoreK         EvalMode = "score-k"
	WeightedScoreK EvalMode = "weighted-score-k"
)

type ModelProjectStats struct {
	Score    float64       `json:"score"`
	Cost     float64       `json:"cost"`
	Duration time.Duration `json:"duration"`
}

func EvalBenchmark(stats BenchmarkStats, mode EvalMode) map[string]map[string]ModelProjectStats {
	scoresPerModelProject := map[string]map[string]ModelProjectStats{}
	for modelName, modelStats := range stats.Models {
		scoresPerModelProject[modelName] = map[string]ModelProjectStats{}
		for projectName, projectStats := range modelStats.Projects {
			stats := ModelProjectStats{}
			for _, reqStats := range projectStats.Requirements {
				stats.Cost += reqStats.Cost
				stats.Duration += reqStats.Duration
				if !reqStats.Completed {
					continue
				}
				switch mode {
				case WeightedScoreK:
					stats.Score += float64(reqStats.MaxScore) / float64(reqStats.Attempts)
				case ScoreK:
					stats.Score += float64(reqStats.MaxScore)
				default:
					panic("invalid eval mode")
				}
			}
			scoresPerModelProject[modelName][projectName] = stats
		}
	}
	return scoresPerModelProject
}

func MustWriteTable(stats map[string]map[string]ModelProjectStats, fpath string, valueFunc func(ModelProjectStats) string) {
	file, err := os.Create(fpath)
	if err != nil {
		panic(err)
	}
	writer := csv.NewWriter(file)
	defer writer.Flush()

	models := []string{}
	for modelName := range stats {
		models = append(models, modelName)
	}

	if len(models) == 0 {
		panic("no models to write")
	}
	sort.Strings(models)

	projects := []string{}
	for projectName := range maps.Keys(stats[models[0]]) {
		projects = append(projects, projectName)
	}
	sort.Strings(projects)

	headers := []string{"project/model"}
	headers = append(headers, models...)
	writer.Write(headers)

	for _, project := range projects {
		row := []string{project}
		for _, model := range models {
			row = append(row, valueFunc(stats[model][project]))
		}
		writer.Write(row)
	}
}
