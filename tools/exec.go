package tools

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type execTool struct {
	allowed map[string]struct{}
	timeout time.Duration
}

func NewExec(allowed []string, timeout time.Duration) Tool {
	commands := make(map[string]struct{}, len(allowed))
	for _, command := range allowed {
		commands[command] = struct{}{}
	}
	if timeout <= 0 {
		timeout = time.Minute
	}
	return execTool{allowed: commands, timeout: timeout}
}

func (execTool) Definition() ToolDefinition {
	return ToolDefinition{Name: "execute_command", Description: "Run one allow-listed command without shell expansion.", Parameters: objectSchema(map[string]any{"command": stringSchema("Command and arguments. Shell operators are not supported.")}, "command")}
}

func (execTool) Safety() Safety {
	return Safety{SideEffect: "shell", Permission: "prompt", Reason: "executes a process on the server"}
}

func (t execTool) Run(ctx context.Context, raw json.RawMessage) Result {
	var input struct {
		Command string `json:"command"`
	}
	if err := json.Unmarshal(raw, &input); err != nil {
		return Result{Content: "invalid command arguments", IsError: true}
	}
	parts := strings.Fields(input.Command)
	if len(parts) == 0 {
		return Result{Content: "command is required", IsError: true}
	}
	if _, ok := t.allowed[parts[0]]; !ok {
		return Result{Content: fmt.Sprintf("command %q is not allow-listed", parts[0]), IsError: true}
	}
	if strings.ContainsAny(input.Command, "|;&><`$") {
		return Result{Content: "shell operators are not supported", IsError: true}
	}
	callCtx, cancel := context.WithTimeout(ctx, t.timeout)
	defer cancel()
	output, err := exec.CommandContext(callCtx, parts[0], parts[1:]...).CombinedOutput()
	result := string(output)
	if len(result) > 8000 {
		result = result[:8000] + "\n... (output truncated)"
	}
	if err != nil {
		return Result{Content: fmt.Sprintf("command failed: %v\n%s", err, result), IsError: true}
	}
	return Result{Content: result}
}
