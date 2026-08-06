package openai_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/AgentHub-Studio/agenthub-go-commons/ai"
	"github.com/AgentHub-Studio/agenthub-go-commons/ai/provider/openai"
)

func TestOpenAIProvider_ChatStream_PreservesRecursiveToolSchema(t *testing.T) {
	parameters := recursiveMetadataFilterParameters()
	var received json.RawMessage
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Tools []struct {
				Function struct {
					Parameters json.RawMessage `json:"parameters"`
				} `json:"function"`
			} `json:"tools"`
		}
		require.NoError(t, json.NewDecoder(r.Body).Decode(&request))
		require.Len(t, request.Tools, 1)
		received = request.Tools[0].Function.Parameters
		w.Header().Set("Content-Type", "text/event-stream")
		_, _ = w.Write([]byte("data: [DONE]\n\n"))
	}))
	t.Cleanup(srv.Close)

	stream, err := openai.New("test-key", srv.URL).ChatStream(
		context.Background(),
		[]ai.Message{{Role: ai.RoleUser, Content: "filter documents"}},
		ai.ChatOptions{
			Model: "gpt-4o",
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
	for chunk := range stream {
		require.NoError(t, chunk.Error)
	}

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
