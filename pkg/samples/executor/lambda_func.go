package executor

import (
	"context"
	"strings"

	"agent-samples/pkg/playbook"
	"agent-samples/pkg/prompt"

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
