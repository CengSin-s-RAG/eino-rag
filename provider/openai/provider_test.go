package openai

import (
	"agent.article.fp/runtime"
	"encoding/json"
	"testing"
)

func TestChatCompletionsToolWireFormat(t *testing.T) {
	request := chatRequest{Model: "test", ToolChoice: "auto", Messages: encodeMessages([]runtime.Message{{Role: runtime.RoleAssistant, ToolCalls: []runtime.ToolCall{{ID: "call_1", Name: "lookup", Arguments: `{"q":"ping"}`}}}}), Tools: encodeTools([]runtime.ToolDefinition{{Name: "lookup", Description: "Lookup a record", Parameters: map[string]any{"type": "object"}}})}
	body, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}
	var decoded map[string]any
	if err := json.Unmarshal(body, &decoded); err != nil {
		t.Fatal(err)
	}
	tool := decoded["tools"].([]any)[0].(map[string]any)
	if tool["type"] != "function" || tool["function"].(map[string]any)["name"] != "lookup" {
		t.Fatalf("unexpected tool payload: %s", body)
	}
	call := decoded["messages"].([]any)[0].(map[string]any)["tool_calls"].([]any)[0].(map[string]any)
	if call["function"].(map[string]any)["arguments"] != `{"q":"ping"}` {
		t.Fatalf("unexpected tool call payload: %s", body)
	}
	if decoded["tool_choice"] != "auto" || decoded["messages"].([]any)[0].(map[string]any)["content"] != nil {
		t.Fatalf("unexpected assistant tool-call payload: %s", body)
	}
}

func TestDecodeToolCall(t *testing.T) {
	message := decodeMessage(wireMessage{Role: runtime.RoleAssistant, ToolCalls: []wireToolCall{{ID: "call_1", Type: "function", Function: wireFunction{Name: "lookup", Arguments: `{"q":"ping"}`}}}})
	if len(message.ToolCalls) != 1 || message.ToolCalls[0].Name != "lookup" {
		t.Fatalf("unexpected decoded message: %#v", message)
	}
}
