package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/adk"
	"github.com/cloudwego/eino/adk/prebuilt/supervisor"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/compose"
	"github.com/cloudwego/eino/schema"
)

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

func buildSearchAgent(ctx context.Context) (adk.Agent, error) {
	m := NewChatModel()

	type searchReq struct {
		Query string `json:"query"`
	}

	type searchResp struct {
		Result string `json:"result"`
	}

	search := func(ctx context.Context, req *searchReq) (*searchResp, error) {
		return &searchResp{
			Result: "In 2024, the US GDP was $29.18 trillion and New York State's GDP was $2.297 trillion",
		}, nil
	}
	searchTool := utils.NewTool(&schema.ToolInfo{
		Name: "search",
		Desc: "search the internet for info",
	}, search)

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "research_agent",
		Description: "the agent responsible to search the internet for info",
		Instruction: `
		You are a research agent.


        INSTRUCTIONS:
        - Assist ONLY with research-related tasks, DO NOT do any math
        - DO NOT estimate any numbers.
        - After you're done with your tasks, respond to the supervisor directly
        - Respond ONLY with the results of your work, do NOT include ANY other text.`,
		Model: m,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{searchTool},
				UnknownToolsHandler: func(ctx context.Context, name, input string) (string, error) {
					return fmt.Sprintf("unknown tool: %s", name), nil
				},
			},
		},
	})
}

func buildMathAgent(ctx context.Context) (adk.Agent, error) {
	m := NewChatModel()

	type addReq struct {
		A float64 `json:"a"`
		B float64 `json:"b"`
	}

	type addResp struct {
		Result float64
	}

	add := func(ctx context.Context, req *addReq) (*addResp, error) {
		return &addResp{
			Result: req.A + req.B,
		}, nil
	}
	addTool, _ := utils.InferTool(
		"add",
		"add two numbers",
		add)

	type multiplyReq struct {
		A float64 `json:"a"`
		B float64 `json:"b"`
	}

	type multiplyResp struct {
		Result float64
	}

	multiply := func(ctx context.Context, req *multiplyReq) (*multiplyResp, error) {
		return &multiplyResp{
			Result: req.A * req.B,
		}, nil
	}
	multiplyTool, _ := utils.InferTool(
		"multiply",
		"multiply two number",
		multiply)

	type divideReq struct {
		A float64 `json:"a"`
		B float64 `json:"b"`
	}

	type divideResp struct {
		Result float64
	}

	divide := func(ctx context.Context, req *divideReq) (*divideResp, error) {
		return &divideResp{
			Result: req.A / req.B,
		}, nil
	}
	divideTool := utils.NewTool(&schema.ToolInfo{
		Name: "divide",
		Desc: "divide two number",
	}, divide)

	return adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "math_agent",
		Description: "the agent responsible to do math",
		Instruction: `
		You are a math agent.


        INSTRUCTIONS:
        - Assist ONLY with math-related tasks
        - After you're done with your tasks, respond to the supervisor directly
        - Respond ONLY with the results of your work, do NOT include ANY other text.`,
		Model: m,
		ToolsConfig: adk.ToolsConfig{
			ToolsNodeConfig: compose.ToolsNodeConfig{
				Tools: []tool.BaseTool{addTool, multiplyTool, divideTool},
				UnknownToolsHandler: func(ctx context.Context, name, input string) (string, error) {
					return fmt.Sprintf("unknown tool: %s", name), nil
				},
			},
		},
	})
}

func buildSupervisor(ctx context.Context) (adk.Agent, error) {
	m := NewChatModel()

	sv, err := adk.NewChatModelAgent(ctx, &adk.ChatModelAgentConfig{
		Name:        "supervisor",
		Description: "the agent responsible to supervise tasks",
		Instruction: `
		You are a supervisor managing two agents:

        - a research agent. Assign research-related tasks to this agent
        - a math agent. Assign math-related tasks to this agent
        Assign work to one agent at a time, do not call agents in parallel.
        Do not do any work yourself.`,
		Model: m,
		Exit:  &adk.ExitTool{},
	})
	if err != nil {
		return nil, err
	}

	searchAgent, err := buildSearchAgent(ctx)
	if err != nil {
		return nil, err
	}
	mathAgent, err := buildMathAgent(ctx)
	if err != nil {
		return nil, err
	}

	return supervisor.New(ctx, &supervisor.Config{
		Supervisor: sv,
		SubAgents:  []adk.Agent{searchAgent, mathAgent},
	})
}

func main() {
	ctx := context.Background()

	sv, err := buildSupervisor(ctx)
	if err != nil {
		log.Fatalf("build supervisor failed: %v", err)
	}

	query := "find US and New York state GDP in 2024. what % of US GDP was New York state?"

	runner := adk.NewRunner(ctx, adk.RunnerConfig{
		Agent:           sv,
		EnableStreaming: true,
	})

	iter := runner.Query(ctx, query)
	fmt.Println("\nuser query: ", query)

	var lastMessage adk.Message
	for {
		event, hasEvent := iter.Next()
		if !hasEvent {
			break
		}

		if event.Output != nil {
			lastMessage, _, err = adk.GetMessage(event)
			fmt.Println(lastMessage.Content)
		}
	}

	// wait for all span to be ended
	time.Sleep(5 * time.Second)

}

func printEvent(evt *adk.AgentEvent) {
	fmt.Printf("name: %s\n", evt.AgentName)
	path := make([]string, 0)
	for _, p := range evt.RunPath {
		path = append(path, p.String())
	}
	fmt.Println(strings.Join(path, ","))
	if evt.Action != nil {
		fmt.Printf("action:%v", *evt.Action)
	} else if evt.Output != nil {
		fmt.Printf("output:%v", evt.Output)
	}

}
