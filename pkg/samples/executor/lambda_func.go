package executor

import (
	"context"
	"strings"

	"agent-samples/pkg/playbook"
	"agent-samples/pkg/prompt"
	"agent-samples/pkg/tool"

	"github.com/cloudwego/eino/schema"
)

// 从playbook提取模板参数
func extractTemplateVariables(ctx context.Context, input playbook.PlayBook) (output map[string]any, err error) {
	variables := map[string]any{}
	variables[prompt.Middleware] = input.Middle

	return variables, nil
}
func collectMessages(ctx context.Context, input []*schema.Message) (output string, err error) {
	strBuilder := strings.Builder{}
	for _, msg := range input {
		strBuilder.WriteString(msg.Content)
	}

	return strBuilder.String(), nil
}

// 调用工具节点
func execTool(ctx context.Context, msg *schema.Message) (map[string]any, error) {
	results := make(map[string]any)
	results[errKey] = make(map[string]string)
	if len(msg.ToolCalls) > 0 {
		for _, call := range msg.ToolCalls {
			t := tool.GetTool(call.Function.Name)
			if t == nil {
				results[errKey].(map[string]string)[call.Function.Name] = "tool not found"
				continue
			}
			result, err := t.InvokableRun(ctx, call.Function.Arguments)
			if err != nil {
				results[errKey].(map[string]string)[call.Function.Name] = err.Error()
			} else {
				results[call.Function.Name] = result
			}
		}
	}

	return results, nil
}

func extractReportVar(ctx context.Context, input *schema.Message) (output map[string]any, err error) {
	output = make(map[string]any)
	if input.Content == finishLabel {
		output[finishLabel] = true
	}
	return output, nil
}
