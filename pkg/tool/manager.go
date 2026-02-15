package tool

import (
	"agent-samples/pkg/tool/impl"
	"context"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/cloudwego/eino/components/tool"
	"gopkg.in/yaml.v3"
)

var toolMap map[string]tool.InvokableTool

// ToolConfigYaml 集群配置
type ToolConfigYaml struct {
	LocalConfigs []impl.ToolConfig `yaml:"local_tools" json:"local_tools"`
	ToolConfigs  []impl.ToolConfig `yaml:"inner_tools" json:"inner_tools"`
}

// InitTool 从配置文件初始化集群配置
func InitTool(configPath string) error {
	// 检查配置文件是否存在
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return fmt.Errorf("配置文件不存在: %s", configPath)
	}

	// 读取配置文件内容
	data, err := ioutil.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("读取配置文件失败: %v", err)
	}

	toolManager := &ToolConfigYaml{}
	toolMap = make(map[string]tool.InvokableTool)
	// 根据文件扩展名决定使用哪种解析方式
	ext := filepath.Ext(configPath)
	switch ext {
	case ".yaml", ".yml":
		// 解析 YAML 配置
		if err := yaml.Unmarshal(data, toolManager); err != nil {
			return fmt.Errorf("解析 YAML 配置文件失败: %v", err)
		}
	case ".json":
		// 解析 JSON 配置
		if err := yaml.Unmarshal(data, toolManager); err != nil {
			return fmt.Errorf("解析 JSON 配置文件失败: %v", err)
		}
	default:
		return fmt.Errorf("不支持的配置文件格式: %s", ext)
	}

	tools, err := toolManager.buildTools()
	if err != nil {
		return err
	}
	ctx := context.TODO()
	for _, tool := range tools {
		info, _ := tool.Info(ctx)
		toolMap[info.Name] = tool
	}

	return nil
}

func GetToolMap() map[string]tool.InvokableTool {
	return toolMap
}

// todo
func (t *ToolConfigYaml) buildTools() ([]tool.InvokableTool, error) {
	tools := make([]tool.InvokableTool, 0)

	return tools, nil
}
