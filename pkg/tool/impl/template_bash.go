package impl

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"text/template"
	"time"

	"github.com/cloudwego/eino/components/tool"
	"github.com/cloudwego/eino/schema"
	"golang.org/x/crypto/ssh"
)

const (
	paramNode = "node"

	authUser     = "username"
	authPassword = "password"
	authKeyPath  = "keyPath"
	authHost     = "host"
	authSSHPort  = "sshPort"

	defaultSSHPort = 22
)

type sshAuth struct {
	User     string
	Password string
	KeyPath  string
	Host     string
	Port     int
}

type TemplateBashTool struct {
	config       *ToolConfig
	execTemplate ExecTemplate
	templateName string
}

func NewTemplateBashTool(cfg *ToolConfig, templateName string) (*TemplateBashTool, error) {
	tmpl, err := findExecTemplate(cfg.ExecTemplates, templateName)
	if err != nil {
		return nil, err
	}

	return &TemplateBashTool{
		config:       cfg,
		execTemplate: tmpl,
		templateName: templateName,
	}, nil
}

func (t *TemplateBashTool) Name() string {
	return t.templateName
}

func (t *TemplateBashTool) Description() string {
	return t.execTemplate.Description
}

func (t *TemplateBashTool) Info(ctx context.Context) (*schema.ToolInfo, error) {
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

	params[paramNode] = &schema.ParameterInfo{
		Type:     "string",
		Desc:     "节点 IP",
		Required: true,
	}

	return &schema.ToolInfo{
		Name:        t.templateName,
		Desc:        t.execTemplate.Description,
		ParamsOneOf: schema.NewParamsOneOfByParams(params),
	}, nil
}

func (t *TemplateBashTool) InvokableRun(ctx context.Context, argumentsInJSON string, opts ...tool.Option) (string, error) {
	args, err := parseArgs(argumentsInJSON)
	if err != nil {
		return "", err
	}

	node := getStringArg(args, paramNode)
	if node == "" {
		return "", errors.New("node is required")
	}

	cmd, err := t.renderCommandTemplate(ctx, t.execTemplate.Exec, args)
	if err != nil {
		return "", err
	}

	return t.executeCommandOnNode(ctx, cmd, node)
}

func (t *TemplateBashTool) renderCommandTemplate(ctx context.Context, tpl string, args map[string]any) (string, error) {
	tmpl, err := template.New("bash").Parse(tpl)
	if err != nil {
		return "", err
	}

	var buf strings.Builder
	if err := tmpl.Execute(&buf, args); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func (t *TemplateBashTool) getSSHClient(ctx context.Context, node string) (*ssh.Client, error) {
	raw, ok := t.config.AuthConfig.NodeAuths[node]
	if !ok {
		return nil, fmt.Errorf("node auth not found: %s", node)
	}

	auth, err := parseSSHAuth(raw)
	if err != nil {
		return nil, err
	}

	cfg, err := buildSSHConfig(auth)
	if err != nil {
		return nil, err
	}

	addr := fmt.Sprintf("%s:%d", auth.Host, auth.Port)

	return ssh.Dial("tcp", addr, cfg)
}

func parseSSHAuth(m map[string]string) (*sshAuth, error) {
	port := defaultSSHPort
	if m[authSSHPort] != "" {
		p, err := strconv.Atoi(m[authSSHPort])
		if err != nil {
			return nil, err
		}
		port = p
	}

	return &sshAuth{
		User:     m[authUser],
		Password: m[authPassword],
		KeyPath:  m[authKeyPath],
		Host:     m[authHost],
		Port:     port,
	}, nil
}

func buildSSHConfig(a *sshAuth) (*ssh.ClientConfig, error) {
	auths := []ssh.AuthMethod{}

	if a.Password != "" {
		auths = append(auths, ssh.Password(a.Password))
	}

	if a.KeyPath != "" {
		key, err := os.ReadFile(a.KeyPath)
		if err != nil {
			return nil, err
		}

		signer, err := ssh.ParsePrivateKey(key)
		if err != nil {
			return nil, err
		}

		auths = append(auths, ssh.PublicKeys(signer))
	}

	return &ssh.ClientConfig{
		User:            a.User,
		Auth:            auths,
		Timeout:         30 * time.Second,
		HostKeyCallback: ssh.InsecureIgnoreHostKey(), // 可后续替换
	}, nil
}

func (t *TemplateBashTool) executeCommandOnNode(ctx context.Context, cmd string, node string) (string, error) {
	client, err := t.getSSHClient(ctx, node)
	if err != nil {
		return "", err
	}
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	var stdout, stderr strings.Builder
	session.Stdout = &stdout
	session.Stderr = &stderr

	if err := session.Run(cmd); err != nil {
		return "", err
	}

	if stderr.Len() > 0 {
		return "", errors.New(stderr.String())
	}

	return stdout.String(), nil
}
