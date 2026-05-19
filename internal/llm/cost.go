package llm

import (
	"embed"
	"encoding/json"
	"strings"
)

//go:embed model_costs.json
var modelCostFS embed.FS

type ModelCost struct {
	InputPerMTokUSD  float64 `json:"input_per_mtok_usd"`
	OutputPerMTokUSD float64 `json:"output_per_mtok_usd"`
}

var modelCosts = loadModelCosts()

func loadModelCosts() map[string]ModelCost {
	data, err := modelCostFS.ReadFile("model_costs.json")
	if err != nil {
		panic(err)
	}
	var costs map[string]ModelCost
	if err := json.Unmarshal(data, &costs); err != nil {
		panic(err)
	}
	return costs
}

// CalculateCost returns the estimated USD cost for a given model and token usage.
// The config stores prices per million tokens. Unknown models cost 0 until added.
func CalculateCost(model string, promptTokens, outputTokens int) float64 {
	cost, ok := modelCost(model)
	if !ok {
		return 0
	}
	return (float64(promptTokens)*cost.InputPerMTokUSD + float64(outputTokens)*cost.OutputPerMTokUSD) / 1_000_000
}

func modelCost(model string) (ModelCost, bool) {
	model = strings.TrimSpace(model)
	if cost, ok := modelCosts[model]; ok {
		return cost, true
	}

	var (
		bestKey  string
		bestCost ModelCost
		found    bool
	)
	for key, cost := range modelCosts {
		if strings.HasPrefix(model, key) && len(key) > len(bestKey) {
			bestKey = key
			bestCost = cost
			found = true
		}
	}
	return bestCost, found
}
