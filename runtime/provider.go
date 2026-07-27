package runtime

import (
	"context"
)

type CompletionRequest struct{}

type Completion struct{}

type Provider interface {
	Complete(ctx context.Context, req CompletionRequest) (Completion, error)
}
