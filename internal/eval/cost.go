package eval

import (
	"fmt"

	"github.com/Sph3ricalPeter/frbench/external"
	"github.com/Sph3ricalPeter/frbench/external/anth"
	"github.com/Sph3ricalPeter/frbench/external/google"
)

type ModelCost struct {
	UsdMtokIn  float64
	UsdMtokOut float64
}

var (
	ModelsCostMap = map[string]ModelCost{
		string(anth.Claude3Haiku): {
			UsdMtokIn:  0.25,
			UsdMtokOut: 1.25,
		},
		string(anth.Claude35Sonnet): {
			UsdMtokIn:  3.0,
			UsdMtokOut: 15.0,
		},
		string(google.Gemini15Flash8B): {
			UsdMtokIn:  0.5,
			UsdMtokOut: 1.5,
		},
		string(google.Gemini15Flash): {
			UsdMtokIn:  0.5,
			UsdMtokOut: 1.5,
		},
		string(google.Gemini15Pro): {
			UsdMtokIn:  1.46,
			UsdMtokOut: 5.87,
		},
	}
)

func MustCalcTotalCost(model string, usage external.ModelUsage) float64 {
	cost, ok := ModelsCostMap[model]
	if !ok {
		panic(fmt.Sprintf("model %s not found in cost map", model))
	}
	return cost.UsdMtokOut*float64(usage.OutputTokens)/1000000 + cost.UsdMtokIn*float64(usage.InputTokens)/1000000
}
