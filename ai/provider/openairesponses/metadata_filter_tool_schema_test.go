package openairesponses

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AgentHub-Studio/agenthub-go-commons/ai"
)

func TestBuildRequest_PreservesRecursiveToolSchema(t *testing.T) {
	parameters := recursiveMetadataFilterParameters()
	req := New("test-key", "").buildRequest(
		[]ai.Message{{Role: ai.RoleUser, Content: "filter documents"}},
		ai.ChatOptions{
			Model: "gpt-5.4",
			Tools: []ai.Tool{{
				Type: "function",
				Function: ai.ToolSchema{
					Name:       "document_search",
					Parameters: parameters,
				},
			}},
		},
		false,
	)

	require.Len(t, req.Tools, 1)
	assert.Equal(t, parameters, req.Tools[0].Parameters)
}

func recursiveMetadataFilterParameters() map[string]any {
	return map[string]any{
		"type": "object",
		"properties": map[string]any{
			"query": map[string]any{"type": "string"},
			"metadataFilter": map[string]any{
				"type": "object",
				"$ref": "#/$defs/metadataFilter",
			},
		},
		"$defs": map[string]any{
			"metadataFilter": map[string]any{
				"oneOf": []any{
					map[string]any{"$ref": "#/$defs/metadataPredicate"},
					map[string]any{"$ref": "#/$defs/metadataAllGroup"},
					map[string]any{"$ref": "#/$defs/metadataAnyGroup"},
					map[string]any{"$ref": "#/$defs/metadataNotGroup"},
				},
			},
		},
	}
}
