package anthropic_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AgentHub-Studio/agenthub-go-commons/ai"
	"github.com/AgentHub-Studio/agenthub-go-commons/ai/provider/anthropic"
)

func TestAnthropicProvider_Chat_PreservesRecursiveToolSchema(t *testing.T) {
	parameters := recursiveMetadataFilterParameters()
	var received json.RawMessage
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Tools []struct {
				InputSchema json.RawMessage `json:"input_schema"`
			} `json:"tools"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		require.Len(t, request.Tools, 1)
		received = request.Tools[0].InputSchema
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]any{
			"id": "msg_metadata_filter", "model": "claude-test", "stop_reason": "end_turn",
			"content": []map[string]any{{"type": "text", "text": "ok"}},
			"usage":   map[string]any{"input_tokens": 1, "output_tokens": 1},
		})
	}))
	t.Cleanup(srv.Close)

	_, err := anthropic.New("test-key", srv.URL).Chat(
		context.Background(),
		[]ai.Message{{Role: ai.RoleUser, Content: "filter documents"}},
		ai.ChatOptions{
			Model: "claude-test",
			Tools: []ai.Tool{{
				Type: "function",
				Function: ai.ToolSchema{
					Name:       "document_search",
					Parameters: parameters,
				},
			}},
		},
	)
	require.NoError(t, err)

	expected, err := json.Marshal(parameters)
	require.NoError(t, err)
	assert.JSONEq(t, string(expected), string(received))
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
