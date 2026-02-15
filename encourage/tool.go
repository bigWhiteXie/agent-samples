package main

import (
	"context"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/components/tool/utils"
	"github.com/cloudwego/eino/schema"
)

type TodoAddParams struct {
	Content  string `json:"content"`
	StartAt  int    `json:"started_at"`
	Deadline int    `json:"deadline"`
}

// 处理函数
func AddTodoFunc(_ context.Context, params *TodoAddParams) (string, error) {
	// Mock处理逻辑
	return `{"msg": "add todo success"}`, nil
}

func getAddTodoTool() tool.InvokableTool {
	// 工具信息
	info := &schema.ToolInfo{
		Name: "add_todo",
		Desc: "Add a todo item",
		ParamsOneOf: schema.NewParamsOneOfByParams(map[string]*schema.ParameterInfo{
			"content": {
				Desc:     "The content of the todo item",
				Type:     schema.String,
				Required: true,
			},
			"started_at": {
				Desc: "The started time of the todo item, in unix timestamp",
				Type: schema.Integer,
			},
			"deadline": {
				Desc: "The deadline of the todo item, in unix timestamp",
				Type: schema.Integer,
			},
		}),
	}

	// 使用NewTool创建工具
	return utils.NewTool(info, AddTodoFunc)
}

type TodoUpdateParams struct {
	ID        string  `json:"id" jsonschema:"description=id of the todo"`
	Content   *string `json:"content,omitempty" jsonschema:"description=content of the todo"`
	StartedAt *int64  `json:"started_at,omitempty" jsonschema:"description=start time in unix timestamp"`
	Deadline  *int64  `json:"deadline,omitempty" jsonschema:"description=deadline of the todo in unix timestamp"`
	Done      *bool   `json:"done,omitempty" jsonschema:"description=done status"`
}

// 处理函数
func UpdateTodoFunc(_ context.Context, params *TodoUpdateParams) (string, error) {
	// Mock处理逻辑
	return `{"msg": "update todo success"}`, nil
}

// 使用 InferTool 创建工具
func GetInferTool() (tool.InvokableTool, error) {
	return utils.InferTool(
		"update_todo", // tool name
		"Update a todo item, eg: content,deadline...", // tool description
		UpdateTodoFunc)
}
