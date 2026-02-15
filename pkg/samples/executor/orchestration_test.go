package executor

import (
	"context"
	"errors"
	"io"
	"testing"

	"agent-samples/pkg/playbook"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBuildplaybook(t *testing.T) {
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
	graph, err := Buildplaybook(ctx, mockPlaybook)
	require.NoError(t, err)
	assert.NotNil(t, graph)
	// invoke调用
	// result, err := graph.Invoke(ctx, *mockPlaybook)
	// assert.NoError(t, err)
	// assert.NotEmpty(t, result)

	// 流式执行
	output, err := graph.Stream(ctx, *mockPlaybook)
	chunk, revErr := output.Recv()
	if revErr != nil {
		if !errors.Is(revErr, io.EOF) {
			t.Error(revErr)
		}
	}
	t.Log(chunk)

}

func TestBuildplaybookWithEmptySteps(t *testing.T) {
	ctx := context.Background()

	// 创建一个空步骤的 mock playbook
	mockPlaybook := &playbook.PlayBook{
		Id:       2,
		Name:     "Empty Steps Playbook",
		TaskGoal: "测试空步骤的目标",
		Middle:   "Test Middleware",
		Steps:    []playbook.Step{},
	}

	// 使用 Buildplaybook 构建图
	graph, err := Buildplaybook(ctx, mockPlaybook)
	require.NoError(t, err)
	assert.NotNil(t, graph)

	// 测试执行图
	result, err := graph.Invoke(ctx, *mockPlaybook)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestBuildplaybookMultipleSteps(t *testing.T) {
	ctx := context.Background()

	// 创建多个步骤的 mock playbook
	steps := make([]playbook.Step, 5)
	for i := 0; i < 5; i++ {
		steps[i] = playbook.Step{
			Name:      "Step " + string(rune(i+'1')),
			Details:   "第 " + string(rune(i+'1')) + " 步详细信息",
			ToolList:  []string{"tool" + string(rune(i+'1'))},
			ToolCalls: []string{},
		}
	}

	mockPlaybook := &playbook.PlayBook{
		Id:       3,
		Name:     "Multiple Steps Playbook",
		TaskGoal: "测试多步骤执行",
		Middle:   "Test Middleware",
		Steps:    steps,
	}

	// 使用 Buildplaybook 构建图
	graph, err := Buildplaybook(ctx, mockPlaybook)
	require.NoError(t, err)
	assert.NotNil(t, graph)

	// 测试执行图
	result, err := graph.Invoke(ctx, *mockPlaybook)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}

func TestBuildplaybookWithHistory(t *testing.T) {
	ctx := context.Background()

	// 创建带有历史记录的 mock playbook
	mockPlaybook := &playbook.PlayBook{
		Id:       4,
		Name:     "Playbook With History",
		TaskGoal: "测试带历史记录的执行",
		Middle:   "Test Middleware",
		Steps: []playbook.Step{
			{
				Name:      "Historical Step",
				Details:   "带历史记录的步骤",
				ToolList:  []string{"history_tool"},
				ToolCalls: []string{},
			},
		},
	}

	// 创建一个 runnable 并测试其执行
	runnable, err := Buildplaybook(ctx, mockPlaybook)
	require.NoError(t, err)
	assert.NotNil(t, runnable)

	// 创建初始状态，包含历史记录
	initialState := playbook.PlayBook{
		Id:       4,
		Name:     "Playbook With History",
		TaskGoal: "测试带历史记录的执行",
		Middle:   "Test Middleware",
		Steps: []playbook.Step{
			{
				Name:      "Historical Step",
				Details:   "带历史记录的步骤",
				ToolList:  []string{"history_tool"},
				ToolCalls: []string{},
			},
		},
	}

	// 执行 runnable
	result, err := runnable.Invoke(ctx, initialState)
	assert.NoError(t, err)
	assert.NotEmpty(t, result)
}
