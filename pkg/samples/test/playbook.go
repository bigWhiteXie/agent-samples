package main

import (
	"agent-samples/pkg/playbook"
	"agent-samples/pkg/samples/executor"
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

func main() {
	ctx := context.Background()

	// 创建 mock 的 playbook
	mockPlaybook := &playbook.PlayBook{
		Id:       1,
		Name:     "Test Playbook",
		TaskGoal: "测试执行目标",
		Middle:   "Test Middleware",
		Steps: []playbook.Step{
			{
				Name:      "Step 1",
				Details:   "第一步详细信息",
				ToolList:  []string{"tool1", "tool2"},
				ToolCalls: []string{},
			},
			{
				Name:      "Step 2",
				Details:   "第二步详细信息",
				ToolList:  []string{"tool3"},
				ToolCalls: []string{},
			},
		},
	}

	// 使用 Buildplaybook 构建图
	graph, err := executor.Buildplaybook(ctx, mockPlaybook)
	if err != nil {
		panic(err)
	}
	// invoke调用
	// result, err := graph.Invoke(ctx, *mockPlaybook)
	// assert.NoError(t, err)
	// assert.NotEmpty(t, result)

	// 流式执行
	msgChan := make(chan *schema.Message, 100)

	// 先启动消息消费 goroutine，避免 channel 阻塞
	done := make(chan struct{})
	go func() {
		defer close(done)
		for msg := range msgChan {
			fmt.Print(msg.Content)
		}
	}()

	callback := executor.NewModelCallback(msgChan)
	output, err := graph.Stream(ctx, *mockPlaybook,
		compose.WithCallbacks(callback).DesignateNode("toolLLM", "analysisLLM", "reportLLM"))
	if err != nil {
		panic(err)
	}

	for {
		chunk, revErr := output.Recv()
		if revErr != nil {
			if !errors.Is(revErr, io.EOF) {
				panic(revErr)
			}
			return
		}
		fmt.Print(chunk)
	}

}
