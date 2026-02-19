package executor

import (
	"agent-samples/pkg/playbook"
	"agent-samples/pkg/prompt"
	"context"

	"github.com/cloudwego/eino/schema"
)

func state2ExecPrompt(ctx context.Context, out map[string]any, state *playbook.State) (map[string]any, error) {
	if out == nil {
		out = make(map[string]any)
	}
	statebook := state.PlayBook

	out[prompt.Middleware] = statebook.Middle
	out[prompt.TaskGoal] = statebook.Steps[state.Current].Details
	out[prompt.ExecutionHistory] = state.History
	out[prompt.Tools] = state.StepCall
	out[prompt.ExecutedTools] = state.CallResult
	out[prompt.ErrorInfo] = state.ErrorInfo
	return out, nil
}

func toolStateHandle(ctx context.Context, out map[string]any, state *playbook.State) (map[string]any, error) {
	result := make(map[string]any)
	// 记录执行错误的工具到state中
	if errInfo, ok := out["errInfo"]; ok {
		if errMap, ok := errInfo.(map[string]string); ok {
			state.ErrorInfo = errMap
		}
	}
	delete(out, "errInfo")
	for toolName, res := range out {
		state.CallResult[toolName] = res.(string)
	}

	// 从待执行工具列表中删除已成功调用的工具
	for name, _ := range out {
		delete(state.StepCall, name)
	}
	result[callKey] = state.StepCall

	return result, nil
}

func state2AnalysisPrompt(ctx context.Context, out map[string]any, state *playbook.State) (map[string]any, error) {
	if out == nil {
		out = make(map[string]any)
	}
	statebook := state.PlayBook

	out[prompt.TaskGoal] = statebook.Steps[state.Current].Details
	out[prompt.ExecutedTools] = state.CallResult
	return out, nil
}

func analysisResultHandle(ctx context.Context, out *schema.Message, state *playbook.State) (*schema.Message, error) {
	state.Current += 1
	// 将分析结果作为当前步骤的诊断结论记录到state中
	state.History = append(state.History, playbook.Record{
		Details: state.PlayBook.Steps[state.Current-1].Details,
		Result:  out.Content,
	})

	if state.Current >= len(state.PlayBook.Steps) {
		out.Content = finishLabel
	} else {
		// 初始化当前步骤待执行工具
		for _, tool := range state.PlayBook.Steps[state.Current].ToolList {
			state.StepCall[tool] = true
		}
	}

	return out, nil
}

func deleteElement[T comparable](arr []T, element ...T) []T {
	result := make([]T, 0)
	for _, v := range arr {
		found := false
		for _, e := range element {
			if v == e {
				found = true
				break
			}
		}
		if !found {
			result = append(result, v)
		}
	}

	return result
}
