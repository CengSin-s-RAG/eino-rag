package util

import (
	"context"
	"fmt"
	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/schema"
)

type SimpleLogger struct{ callbacks.HandlerBuilder }

func (l *SimpleLogger) OnStart(ctx context.Context, info *callbacks.RunInfo, input callbacks.CallbackInput) context.Context {
	return ctx
}

func (l *SimpleLogger) OnEnd(ctx context.Context, info *callbacks.RunInfo, output callbacks.CallbackOutput) context.Context {
	fmt.Println(output)
	return ctx
}

func (l *SimpleLogger) OnEndWithStreamOutput(ctx context.Context, info *callbacks.RunInfo, output *schema.StreamReader[callbacks.CallbackOutput]) context.Context {
	output.Close()
	return ctx
}

func (l *SimpleLogger) OnStartWithStreamInput(ctx context.Context, info *callbacks.RunInfo, input *schema.StreamReader[callbacks.CallbackInput]) context.Context {
	input.Close()
	return ctx
}

func (l *SimpleLogger) OnError(ctx context.Context, info *callbacks.RunInfo, err error) context.Context {
	fmt.Println(err)
	return ctx
}
