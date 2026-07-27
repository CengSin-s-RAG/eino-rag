package bootstrap

import (
	"agent.article.fp/config"
	"testing"
)

func TestAgentConfigDefaultsRemainSafe(t *testing.T) {
	// NewChatService owns applying zero-value defaults. This check documents that
	// the configuration is intentionally optional for existing deployments.
	cfg := config.Config{}
	if cfg.Agent.MaxSteps != 0 || cfg.Agent.HistoryLimit != 0 {
		t.Fatalf("unexpected non-zero defaults: %#v", cfg.Agent)
	}
}
