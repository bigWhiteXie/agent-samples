package executor

import (
	"context"
	"log"

	"agent-samples/pkg/playbook"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

const (
	errKey  = "errInfo"
	callKey = "callInfo"

	finishLabel = "finish"

	promptVarNode        = "PromptVarNode"
	templateNode         = "templateNode"
	toolLLM              = "toolLLM"
	execToolNode         = "execToolNode"
	analysisVarNode      = "analysisVarNode"
	analysisTemplateNode = "analysisTemplateNode"
	analysisLLM          = "analysisLLM"
	reportVarNode        = "reportVarNode"
	reportTemplateNode   = "reportTemplateNode"
	reportLLM            = "reportLLM"
	Playbook             = "playbook"
)

func Buildplaybook(ctx context.Context, book *playbook.PlayBook) (r compose.Runnable[playbook.PlayBook, *schema.Message], err error) {
	g := compose.NewGraph[playbook.PlayBook, *schema.Message](compose.WithGenLocalState(func(ctx context.Context) (state *playbook.State) {
		callmap := make(map[string]bool)
		if len(book.Steps) > 0 {
			for _, tool := range book.Steps[0].ToolList {
				callmap[tool] = true
			}
		}
		return &playbook.State{
			Current:    0,
			PlayBook:   book,
			History:    make([]playbook.Record, 0),
			StepCall:   callmap,
			CallResult: make(map[string]string),
			ErrorInfo:  make(map[string]string),
		}
	}))
	// 阶段1 工具诊断: 根据当前步骤、执行历史、工具调用信息调用工具
	// 从playbook、state中提取上下文参数
	g.AddLambdaNode(promptVarNode, compose.InvokableLambda(extractTemplateVariables))
	// 构建上下文节点
	templateNodeKeyOfChatTemplate, err := newChatTemplate(ctx)
	if err != nil {
		return nil, err
	}
	g.AddChatTemplateNode(templateNode, templateNodeKeyOfChatTemplate, compose.WithStatePreHandler(state2ExecPrompt))
	// 构建llm节点和执行工具节点
	chatModel := NewChatModelByBook(book)
	g.AddChatModelNode(toolLLM, chatModel)
	g.AddLambdaNode(execToolNode, compose.InvokableLambda(execTool), compose.WithStatePostHandler(toolStateHandle)) //输出map[string]any,在post钩子更新state的异常信息、工具调用结果、涉及工具列表
	// 分叉节点，判断是否还存在剩余工具，不存在则分析当前步骤执行结果
	br1 := compose.NewGraphBranch(func(ctx context.Context, in map[string]any) (endNode string, err error) {
		if callInfo, ok := in[callKey]; ok {
			if callList, ok := callInfo.(map[string]bool); ok {
				if len(callList) > 0 {
					// 重新返回template节点去执行
					return templateNode, nil
				}
			}
		}
		// 返回下一阶段的template节点
		return analysisTemplateNode, nil
	}, map[string]bool{templateNode: true, analysisTemplateNode: true})

	// 阶段2 调用结果分析：根据调用结果、历史记录提炼当前步骤的诊断结果
	// 构建llm用于分析当前阶段的执行过程，提取出精练的执行结果
	analysisTemplate, err := newAnalaysisChatTemplate(ctx)
	g.AddChatTemplateNode(analysisTemplateNode, analysisTemplate, compose.WithStatePreHandler(state2AnalysisPrompt))
	if err != nil {
		return nil, err
	}
	analysisModel := NewChatModel()
	// 更新state中的curindex和历史执行结果
	g.AddChatModelNode(analysisLLM, analysisModel, compose.WithStatePostHandler(analysisResultHandle))

	// 阶段三：根据执行记录生成诊断报告
	g.AddLambdaNode(reportVarNode, compose.InvokableLambda(extractReportVar))
	// 更新完后若没有剩余步骤则进入诊断报告生成阶段
	br2 := compose.NewGraphBranch(func(ctx context.Context, in map[string]any) (endNode string, err error) {
		if _, ok := in[finishLabel]; ok {
			return reportTemplateNode, nil
		}
		return templateNode, nil
	}, map[string]bool{templateNode: true, reportTemplateNode: true})
	reportTemplate, err := newReportChatTemplate(ctx)
	if err != nil {
		return nil, err
	}
	g.AddChatTemplateNode(reportTemplateNode, reportTemplate, compose.WithStatePreHandler(state2ExecPrompt))
	reportModel := NewChatModel()
	g.AddChatModelNode(reportLLM, reportModel)

	// 构建图、添加节点和边
	_ = g.AddEdge(compose.START, promptVarNode)
	_ = g.AddEdge(promptVarNode, templateNode)
	_ = g.AddEdge(templateNode, toolLLM)
	_ = g.AddEdge(toolLLM, execToolNode)
	g.AddBranch(execToolNode, br1)
	_ = g.AddEdge(analysisTemplateNode, analysisLLM)
	_ = g.AddEdge(analysisLLM, reportVarNode)
	g.AddBranch(reportVarNode, br2)
	_ = g.AddEdge(reportTemplateNode, reportLLM)
	_ = g.AddEdge(reportLLM, compose.END)

	r, err = g.Compile(ctx, compose.WithGraphName("playbook"), compose.WithMaxRunSteps(100))
	if err != nil {
		return nil, err
	}

	return r, err
}

func NewChatModelByBook(book *playbook.PlayBook) model.ToolCallingChatModel {
	cm, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		Model:   "qwen3",
		BaseURL: "http://127.0.0.1:11434/v1",
	})
	if err != nil {
		log.Fatal(err)
	}

	toolInfos := make([]*schema.ToolInfo, 0)
	tools := book.GetTools()
	for _, t := range tools {
		toolInfo, err := t.Info(context.TODO())
		if err != nil {
			log.Printf("获取工具信息失败: %v", err)
			continue
		}
		toolInfos = append(toolInfos, toolInfo)
	}
	cm.WithTools(toolInfos)

	return cm
}

func NewChatModel() model.ToolCallingChatModel {
	cm, err := openai.NewChatModel(context.Background(), &openai.ChatModelConfig{
		Model:   "qwen3",
		BaseURL: "http://127.0.0.1:11434/v1",
	})
	if err != nil {
		log.Fatal(err)
	}

	return cm
}
