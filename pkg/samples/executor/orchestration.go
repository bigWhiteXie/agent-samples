package executor

import (
	"context"
	"fmt"
	"strings"

	"agent-samples/pkg/playbook"
	"agent-samples/pkg/prompt"

	"github.com/cloudwego/eino/compose"
)

func Buildplaybook(ctx context.Context, book *playbook.PlayBook) (r compose.Runnable[playbook.PlayBook, string], err error) {
	const (
		PromptVarNode = "PromptVarNode"
		templateNode  = "templateNode"
		CollectMsg    = "CollectMsg"
		Playbook      = "playbook"
	)
	g := compose.NewGraph[playbook.PlayBook, string](compose.WithGenLocalState(func(ctx context.Context) (state *playbook.State) {
		return &playbook.State{
			Current:  0,
			PlayBook: book,
		}
	}))

	_ = g.AddLambdaNode(PromptVarNode, compose.InvokableLambda(extractTemplateVariables),
		compose.WithStatePostHandler(func(ctx context.Context, out map[string]any, state *playbook.State) (map[string]any, error) {
			statebook := state.PlayBook
			out[prompt.TaskGoal] = fmt.Sprintf("执行:%s\n涉及工具:%s", statebook.Steps[state.Current].Details, strings.Join(statebook.Steps[state.Current].ToolList, ","))
			out[prompt.ExecutionHistory] = state.History

			return out, nil
		}))
	templateNodeKeyOfChatTemplate, err := newChatTemplate(ctx)
	if err != nil {
		return nil, err
	}

	_ = g.AddChatTemplateNode(templateNode, templateNodeKeyOfChatTemplate)
	_ = g.AddLambdaNode(CollectMsg, compose.InvokableLambda(collectMessages))
	_ = g.AddEdge(compose.START, PromptVarNode)
	_ = g.AddEdge(CollectMsg, compose.END)
	_ = g.AddEdge(PromptVarNode, templateNode)
	_ = g.AddEdge(templateNode, CollectMsg)
	r, err = g.Compile(ctx, compose.WithGraphName("playbook"), compose.WithNodeTriggerMode(compose.AllPredecessor))
	if err != nil {
		return nil, err
	}

	return r, err
}
