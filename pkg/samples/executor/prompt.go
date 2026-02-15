package executor

import (
	"context"

	inprompt "agent-samples/pkg/prompt"

	"github.com/cloudwego/eino/components/prompt"
	"github.com/cloudwego/eino/schema"
)

// newChatTemplate component initialization function of node 'templateNode' in graph 'playbook'
func newChatTemplate(ctx context.Context) (ctp prompt.ChatTemplate, err error) {
	ctp = prompt.FromMessages(schema.GoTemplate,
		&schema.Message{
			Role:    schema.System,
			Content: inprompt.SystemPlaybook,
		},
		&schema.Message{
			Role:    schema.User,
			Content: inprompt.UserPlaybook,
		},
	)
	return ctp, nil
}
