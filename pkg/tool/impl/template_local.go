package impl

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
)

type TemplateLocalTool struct {
	config       *ToolConfig
	execTemplate ExecTemplate
	templateName string
}

func NewTemplateLocalTool(cfg *ToolConfig, templateName string) (*TemplateLocalTool, error) {
	tmpl, err := findExecTemplate(cfg.ExecTemplates, templateName)
	if err != nil {
		return nil, err
	}

	return &TemplateLocalTool{
		config:       cfg,
		execTemplate: tmpl,
		templateName: templateName,
	}, nil
}

func (t *TemplateLocalTool) Name() string {
	return t.templateName
}

func (t *TemplateLocalTool) Description() string {
	return t.execTemplate.Description
}

func (t *TemplateLocalTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
	params := make(map[string]*schema.ParameterInfo)

	for _, p := range t.execTemplate.Parameters {
		info := &schema.ParameterInfo{
			Type:     "string",
			Desc:     p.Description,
			Required: p.Required,
			Enum:     p.Enum,
		}
		params[p.Name] = info
	}

	return &schema.ToolInfo{
		Name:        t.templateName,
		Desc:        t.execTemplate.Description,
		ParamsOneOf: schema.NewParamsOneOfByParams(params),
	}, nil
}

func (t *TemplateLocalTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	args, err := parseArgs(argumentsInJSON)
	if err != nil {
		return "", err
	}

	node := getStringArg(args, paramNode)
	if node == "" {
		return "", errors.New("node is required")
	}

	cmd, err := renderCommandTemplate("local", t.execTemplate.Exec, args)
	if err != nil {
		return "", err
	}

	// 使用 os/exec 包在当前环境执行命令
	cmdObj := exec.CommandContext(ctx, "/bin/sh", "-c", cmd)

	var stdout, stderr bytes.Buffer
	cmdObj.Stdout = &stdout
	cmdObj.Stderr = &stderr

	if err := cmdObj.Run(); err != nil {
		return "", fmt.Errorf("error executing command: %v, stderr: %s", err, stderr.String())
	}

	if stderr.Len() > 0 {
		return "", errors.New(stderr.String())
	}

	return stdout.String(), nil
}
