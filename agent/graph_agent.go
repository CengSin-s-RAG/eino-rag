package agent

import (
	"agent.article.fp/agent/component"
	"agent.article.fp/client"
	"agent.article.fp/config"
	"agent.article.fp/skill"
	"agent.article.fp/util"
	"context"
	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/flow/agent/react"
	"github.com/cloudwego/eino/schema"
)

// EinoChatAgent 封装编译好的 Runnable，对外提供服务
type EinoChatAgent struct {
	Runner       *react.Agent
	SkillCatalog string // Skill 目录文本，注入系统提示词
}

func NewEinoChatAgent(ctx context.Context, modelConf openai.ChatModelConfig) (*EinoChatAgent, error) {
	// 初始化 Skill Manager
	skillMgr, err := skill.NewManager(config.Cfg.SkillsPath)
	if err != nil {
		return nil, err
	}

	// 如果有可用 Skill，注册 activate_skill 工具
	var skillCatalog string
	if skillMgr.HasSkills() {
		skillTool := client.NewActivateSkillTool(skillMgr)
		info, err := skillTool.Info(ctx)
		if err != nil {
			return nil, err
		}
		client.ToolsInfo = append(client.ToolsInfo, info)
		client.EinoTools = append(client.EinoTools, skillTool)
		skillCatalog = skillMgr.Catalog()
	}

	reactAgent, err := newReactLambdaAgent(ctx, modelConf, skillCatalog)
	if err != nil {
		return nil, err
	}
	return &EinoChatAgent{Runner: reactAgent, SkillCatalog: skillCatalog}, nil
}

func newReactLambdaAgent(ctx context.Context, cfg openai.ChatModelConfig, skillCatalog string) (*react.Agent, error) {
	hybirdRetriever, err := component.BuildContentRetriever(ctx, &component.MyContentRetrieverConfig{
		ColName:    util.CollectionFupengshuo,
		Retrievers: []string{"vectors", "full_text"},
	})
	if err != nil {
		return nil, err
	}

	// 1. 定义 Chat Model (复用你的配置)
	chatModel, err := openai.NewChatModel(ctx, &cfg)
	if err != nil {
		return nil, err
	}
	kbTool := client.NewRetrieverTool(hybirdRetriever,
		"search_financial_knowledge",
		"Use this tool to search for internal financial reports, news, and articles.",
	)
	info, err := kbTool.Info(ctx)
	if err != nil {
		return nil, err
	}
	client.ToolsInfo = append(client.ToolsInfo, info)
	client.EinoTools = append(client.EinoTools, kbTool)
	if err = chatModel.BindTools(client.ToolsInfo); err != nil {
		return nil, err
	}

	agentConfig := react.AgentConfig{
		ToolCallingModel: chatModel,
		ToolsConfig:      compose.ToolsNodeConfig{Tools: client.EinoTools},
		MaxStep:          6,
		MessageModifier: func(ctx context.Context, input []*schema.Message) []*schema.Message {
			// 使用小模型进行摘要
			if len(input) > 19 && schema.User == input[len(input)-1].Role {
				newQuery, err := component.QueryRewriting(input[1:])
				if err != nil {
					return input
				}
				return []*schema.Message{schema.SystemMessage(util.GetSystemPrompt(skillCatalog)), schema.UserMessage(newQuery)}
			}
			return input
		},
	}

	agent, err := react.NewAgent(ctx, &agentConfig)
	if err != nil {
		return nil, err
	}

	return agent, nil
}
