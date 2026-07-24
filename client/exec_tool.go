package client

import (
	"agent.article.fp/config"
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
)

// NewExecCommandTool 创建 execute_command 工具，允许 LLM 执行白名单内的 Shell 命令
func NewExecCommandTool() tool.BaseTool {
	runFunc := func(ctx context.Context, input struct {
		Command string `json:"command" jsonschema:"description=The shell command to execute"`
	}) (string, error) {
		cmdStr := strings.TrimSpace(input.Command)
		if cmdStr == "" {
			return "[command rejected: empty command]", nil
		}

		// 提取命令的第一部分（可执行文件名）进行白名单校验
		if err := validateCommand(cmdStr); err != nil {
			return fmt.Sprintf("[command rejected: %s]", err), nil
		}

		// 设置超时
		timeout := 60
		if config.Cfg.Exec != nil && config.Cfg.Exec.Timeout > 0 {
			timeout = config.Cfg.Exec.Timeout
		}
		execCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()

		// 执行命令
		cmd := exec.CommandContext(execCtx, "sh", "-c", cmdStr)
		output, err := cmd.CombinedOutput()

		result := string(output)

		// 截断过长的输出（避免 token 爆炸）
		const maxOutputLen = 8000
		if len(result) > maxOutputLen {
			result = result[:maxOutputLen] + "\n... (output truncated)"
		}

		if err != nil {
			if execCtx.Err() == context.DeadlineExceeded {
				return fmt.Sprintf("[command timed out after %d seconds]\n%s", timeout, result), nil
			}
			return fmt.Sprintf("[command exited with error: %s]\n%s", err, result), nil
		}

		return result, nil
	}

	newTool, _ := utils.InferTool(
		"execute_command",
		"Execute a shell command on the server. Only whitelisted commands are allowed. Use for data processing, API calls, or running scripts.",
		runFunc,
	)
	return newTool
}

// validateCommand 校验命令是否在白名单中
func validateCommand(cmdStr string) error {
	if config.Cfg.Exec == nil || len(config.Cfg.Exec.AllowedCommands) == 0 {
		return fmt.Errorf("command execution is disabled: no allowed commands configured")
	}

	// 提取命令名（第一个空格前的部分，处理管道和链接符）
	// 也处理 /usr/bin/ls 这种绝对路径
	parts := strings.Fields(cmdStr)
	if len(parts) == 0 {
		return fmt.Errorf("empty command")
	}

	firstCmd := parts[0]
	// 如果是绝对路径，取最后一段作为命令名
	if idx := strings.LastIndex(firstCmd, "/"); idx >= 0 {
		firstCmd = firstCmd[idx+1:]
	}

	allowed := config.Cfg.Exec.AllowedCommands
	for _, a := range allowed {
		if firstCmd == a {
			return nil
		}
	}

	return fmt.Errorf("command '%s' is not in the allowed commands list: %v", firstCmd, allowed)
}
