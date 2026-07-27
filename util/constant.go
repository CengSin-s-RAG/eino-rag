package util

import (
	_ "embed"
)

var (
	//go:embed systemPrompt.md
	SystemPrompt string

	//go:embed rerankSystemPrompt.txt
	RerankPrompt string

	//go:embed rewriteSystemPrompt.txt
	QueryRewritePrompt string
)
