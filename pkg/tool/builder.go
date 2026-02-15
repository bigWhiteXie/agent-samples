package tool

import (
	"agent-samples/pkg/tool/impl"

	"github.com/cloudwego/eino/components/tool"
)

type Builder func(toolConfig ToolConfigYaml) []tool.InvokableTool

func BuildBashTool(toolConfig ToolConfigYaml) []tool.InvokableTool {
	var bashTools []tool.InvokableTool

	// 遍历所有工具配置，查找名为"bash"的工具
	for _, config := range toolConfig.ToolConfigs {
		if config.ToolName == "bash" {
			// 对于bash工具中的每个执行模板，创建一个TemplateBashTool实例
			for _, execTemplate := range config.ExecTemplates {
				bashTool, err := impl.NewTemplateBashTool(&config, execTemplate.Name)
				if err != nil {
					// 如果创建失败，跳过这个模板
					continue
				}
				bashTools = append(bashTools, bashTool)
			}
		}
	}

	return bashTools
}
