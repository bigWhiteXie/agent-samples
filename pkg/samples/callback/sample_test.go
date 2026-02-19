package callback

import (
	"context"
	"testing"

	"github.com/cloudwego/eino/callbacks"
	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func TestSample(t *testing.T) {
	ctx := context.Background()
	top := compose.NewGraph[map[string]any, []*schema.Message]()
	sub := compose.NewGraph[map[string]any, []*schema.Message]()
	_ = sub.AddChatTemplateNode("tmpl_nested", prompt.FromMessages(schema.FString, schema.UserMessage("Hello, {name}!")))
	_ = sub.AddEdge(compose.START, "tmpl_nested")
	_ = sub.AddEdge("tmpl_nested", compose.END)
	_ = top.AddGraphNode("sub_graph", sub)
	_ = top.AddEdge(compose.START, "sub_graph")
	_ = top.AddEdge("sub_graph", compose.END)
	r, _ := top.Compile(ctx)

	optGlobal := compose.WithCallbacks(
		callbacks.NewHandlerBuilder().OnEndFn(func(ctx context.Context, _ *callbacks.RunInfo, _ callbacks.CallbackOutput) context.Context {
			return ctx
		}).Build(),
	)
	optNode := compose.WithCallbacks(
		callbacks.NewHandlerBuilder().OnStartFn(func(ctx context.Context, _ *callbacks.RunInfo, _ callbacks.CallbackInput) context.Context { return ctx }).Build(),
	).DesignateNode("sub_graph")
	optNested := compose.WithChatTemplateOption(
		prompt.WrapImplSpecificOptFn(func(_ *struct{}) {}),
	).DesignateNodeWithPath(
		compose.NewNodePath("sub_graph", "tmpl_nested"),
	)

	_, _ = r.Invoke(ctx, map[string]any{"name": "Alice"}, optGlobal, optNode, optNested)
}
