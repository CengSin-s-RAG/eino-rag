// Package bootstrap is the application manifest. It is the explicit, testable
// equivalent of V2Ray's distro/all blank-import manifest.
package bootstrap

import (
	"agent.article.fp/config"
	"agent.article.fp/provider/openai"
	"agent.article.fp/runtime"
	"agent.article.fp/service"
	"agent.article.fp/skill"
	"agent.article.fp/store"
	"agent.article.fp/tools"
	"agent.article.fp/util"
	"context"
	"fmt"
	"time"
)

type App struct {
	Runtime *runtime.App
	Chat    *service.ChatService
	Aux     *service.AuxiliaryService
}

func New(ctx context.Context, cfg *config.Config) (*App, error) {
	if cfg == nil {
		return nil, fmt.Errorf("config is required")
	}
	primary, err := openai.New(cfg.Providers.Primary)
	if err != nil {
		return nil, fmt.Errorf("create primary provider: %w", err)
	}
	fallback := make([]runtime.Provider, 0, len(cfg.Providers.Fallback))
	closers := []runtime.Closer{primary}
	for _, providerCfg := range cfg.Providers.Fallback {
		provider, err := openai.New(providerCfg)
		if err != nil {
			return nil, fmt.Errorf("create fallback provider: %w", err)
		}
		fallback = append(fallback, provider)
		closers = append(closers, provider)
	}

	chatStore := store.NewRedis(*cfg.Redis)
	closers = append(closers, chatStore)
	skillManager, err := skill.NewManager(cfg.SkillsPath)
	if err != nil {
		return nil, err
	}
	registered := []tools.Tool{tools.NewSkill(skillManager)}
	if len(cfg.MCP) == 0 {
		return nil, fmt.Errorf("at least one MCP service must be configured")
	}
	mcpCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()
	for _, mcpConfig := range cfg.MCP {
		mcpTools, mcpClient, err := tools.NewMCP(mcpCtx, mcpConfig)
		if err != nil {
			return nil, err
		}
		registered = append(registered, mcpTools...)
		closers = append(closers, mcpClient)
	}
	if cfg.Exec != nil && len(cfg.Exec.AllowedCommands) > 0 {
		registered = append(registered, tools.NewExec(cfg.Exec.AllowedCommands, time.Duration(cfg.Exec.Timeout)*time.Second))
	}
	registry, err := tools.NewRegistry(registered...)
	if err != nil {
		return nil, err
	}
	closers = append(closers, registry)

	chat := service.NewChatService(
		primary,
		fallback,
		registry,
		chatStore,
		util.GetSystemPrompt(skillManager.Catalog()),
		cfg.Agent.HistoryLimit,
		cfg.Agent.MaxSteps,
	)
	return &App{Runtime: runtime.NewApp(closers...), Chat: chat, Aux: service.NewAuxiliaryService(primary, chatStore)}, nil
}

func (a *App) Close() error { return a.Runtime.Close() }
