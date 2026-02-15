package impl

import (
	"encoding/json"
	"fmt"
	"strings"
	"text/template"
)

type AuthType string

const (
	AuthTypeNone    AuthType = "none"    // 无需认证
	AuthTypeGlobal  AuthType = "global"  // 全局认证
	AuthTypePerNode AuthType = "perNode" // 每个节点单独认证
)

// AuthTypeDescriptions 认证类型描述映射
var AuthTypeDescriptions = map[AuthType]string{
	AuthTypeNone:    "该工具无需认证和节点信息，直接执行即可",
	AuthTypeGlobal:  "该工具无需节点信息，但是需要传入认证信息",
	AuthTypePerNode: "该工具涉及到多节点操作，并且区分每个节点的认证信息",
}

// GetAuthTypeDescription 获取认证类型的描述
func GetAuthTypeDescription(authType AuthType) (string, error) {
	if description, exists := AuthTypeDescriptions[authType]; exists {
		return description, nil
	}
	return "", fmt.Errorf("未知的认证类型")
}

// AuthConfig 认证配置
type AuthConfig struct {
	Type       AuthType                     `json:"type"`       // 认证类型
	GlobalAuth map[string]string            `json:"globalAuth"` // 全局认证信息，所有节点通用
	NodeAuths  map[string]map[string]string `json:"nodeAuths"`  // 节点级别认证信息，key为节点名称
}

// ExecTemplate 执行模板配置
type ExecTemplate struct {
	Name        string      `json:"name"`                         // 标识符
	Description string      `json:"description"`                  // 描述，用于解释该命令的作用和使用方法
	Exec        string      `json:"exec"`                         // 执行模板，可以使用模板参数
	Parameters  []Parameter `json:"parameters" yaml:"parameters"` // 参数列表
}

// Parameter 参数定义
type Parameter struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Required    bool     `json:"required" yaml:"required"`
	Enum        []string `json:"enum,omitempty" yaml:"enum,omitempty"` // 可选值列表
}

// 自定义工具配置结构体
type CustomTool struct {
	ToolName      string         `json:"toolName" yaml:"toolName"`
	ToolDesc      string         `json:"description" yaml:"description"`
	AuthConfig    *AuthConfig    `json:"authConfig" yaml:"authConfig"`       // 认证配置
	ExecTemplates []ExecTemplate `json:"execTemplates" yaml:"execTemplates"` // 执行模板表
	Timeout       int            `json:"timeout" yaml:"timeout"`             // 超时时间（秒），默认30秒
}

// 工具配置结构体
type ToolConfig struct {
	ToolName      string         `json:"toolName" yaml:"toolName"` // 工具名称
	ToolDesc      string         `json:"description" yaml:"description"`
	AuthConfig    *AuthConfig    `json:"authConfig" yaml:"authConfig"`       // 认证配置
	ExecTemplates []ExecTemplate `json:"execTemplates" yaml:"execTemplates"` // 执行模板表

	Extra map[string]string `json:"extra" yaml:"extra"` // 额外信息
}

func parseArgs(raw string) (map[string]any, error) {
	var m map[string]any
	if err := json.Unmarshal([]byte(raw), &m); err != nil {
		return nil, fmt.Errorf("invalid arguments json: %w", err)
	}
	return m, nil
}

func getStringArg(m map[string]any, key string) string {
	if v, ok := m[key]; ok {
		if s, ok := v.(string); ok {
			return s
		}
	}
	return ""
}

func findExecTemplate(templates []ExecTemplate, name string) (ExecTemplate, error) {
	for _, t := range templates {
		if t.Name == name {
			return t, nil
		}
	}
	return ExecTemplate{}, fmt.Errorf("exec template not found: %s", name)
}

func renderCommandTemplate(name, tpl string, args map[string]any) (string, error) {
	tmpl, err := template.New(name).Parse(tpl)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, args); err != nil {
		return "", err
	}

	return buf.String(), nil
}
