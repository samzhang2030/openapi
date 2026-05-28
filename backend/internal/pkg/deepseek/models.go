package deepseek

import "github.com/Wei-Shaw/sub2api/internal/pkg/openai"

// DefaultModels is the curated DeepSeek model list exposed by /v1/models.
var DefaultModels = []openai.Model{
	{ID: "deepseek-v4-pro", Object: "model", Created: 1704067200, OwnedBy: "deepseek", Type: "model", DisplayName: "DeepSeek V4 Pro"},
	{ID: "deepseek-v4-flash", Object: "model", Created: 1704067200, OwnedBy: "deepseek", Type: "model", DisplayName: "DeepSeek V4 Flash"},
}

// DefaultTestModel is the model used for account connectivity tests.
const DefaultTestModel = "deepseek-v4-pro"

// DefaultModelIDs returns the default DeepSeek model ID list.
func DefaultModelIDs() []string {
	ids := make([]string, len(DefaultModels))
	for i, model := range DefaultModels {
		ids[i] = model.ID
	}
	return ids
}
