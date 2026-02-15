package playbook

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bytedance/gopkg/util/logger"
)

type Step struct {
	Name      string   `json:"name" yaml:"name"`
	Details   string   `json:"details" yaml:"details"`
	ToolList  []string `json:"tool_list" yaml:"tool_list"`
	ToolCalls []string
}

func (s Step) GetToolNames() string {
	return strings.Join(s.ToolList, ",")
}

type Record struct {
	Details string
	Result  string
}

type State struct {
	Current  int      // 当前执行步骤
	History  []Record // 历史执行结果
	PlayBook *PlayBook
}

// 运维诊断方案
type PlayBook struct {
	Id       int    `json:"id" `
	Name     string `json:"name" yaml:"name"`
	TaskGoal string `json:"task_goal" yaml:"task_goal"`
	Middle   string `json:"middle" yaml:"middle"`
	Steps    []Step `json:"steps" yaml:"steps"`
	Details  string `json:"details" ` // 存放steps的序列化内容
}

// GetCollection 返回集合名称
func (p *PlayBook) GetCollection() string {
	return "playbook_collection"
}

// SetPartition 设置分区
func (p *PlayBook) SetPartition(partition string) {
	p.Middle = partition
}

// GetPartition 获取分区
func (p *PlayBook) GetPartition() string {
	return ""
}

func (p *PlayBook) GetMiddleName() string {
	return p.Middle
}

func (p *PlayBook) SetDetails() {
	if len(p.Steps) != 0 {
		details, err := json.Marshal(p.Steps)
		if err != nil {
			logger.Errorf("fail to marshal steps to detais:%s", err)
			return
		}
		p.Details = string(details)
	}
}

// Format 将steps转换为details格式
// 格式：{{步骤名称}}: {{具体信息}}
func (p *PlayBook) Format() string {
	if len(p.Steps) == 0 {
		p.Details = ""
		return ""
	}

	var detailsBuilder strings.Builder
	detailsBuilder.WriteString(fmt.Sprintf("方案名称:%s\n", p.Name))
	detailsBuilder.WriteString(fmt.Sprintf("方案目标:%s\n", p.TaskGoal))
	for i, step := range p.Steps {
		detailsBuilder.WriteString(fmt.Sprintf("步骤%d:%s\n %s\n\n", i+1, step.Name, step.Details))
	}

	return detailsBuilder.String()
}
