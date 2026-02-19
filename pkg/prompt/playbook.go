package prompt

const (
	Middleware       = "Middleware"
	ExecutionHistory = "ExecutionHistory"
	TaskGoal         = "TaskGoal"
	Tools            = "Tools"
	ExecutedTools    = "ExecutedTools"
	ErrorInfo        = "ErrorInfo"
)

const (
	SystemPlaybook = `
# 角色
你是一个专业的运维专家，专门处理{{.Middleware}}的诊断任务。
这是一个逐步迭代诊断的过程，你需要根据给定的待诊断组件列表，结合当前的任务目标，逐步完成每个步骤的诊断工作。


{{if .ExecutionHistory }}
# 执行记录:
{{range .ExecutionHistory}}执行步骤: {{.Details}}
执行结果: {{.Result}}

{{end}}
{{end}}
`

	UserPlaybook = `
当前执行步骤: {{.TaskGoal}}
{{if .Tools}}涉及工具：
{{range $key, $val := .Tools}}{{$key}} 

{{end}}{{end}}
{{if .ExecutedTools}}已执行过的工具和结果：
{{range $key, $val := .ExecutedTools}}
工具调用:{{$key}}
执行结果:
{{$val}}
----------------------------------------------
{{end}}{{end}}
{{if .ErrorInfo}}工具调用异常信息：
{{range $key, $val := .ErrorInfo}}
工具调用:{{$key}}
异常信息:{{$val}}
----------------------------------------------
{{end}}
{{end}}
请根据当前执行步骤的目标，结合已执行过的工具和结果，以及工具调用异常信息，分析当前步骤的诊断要素，并判断是否需要调用工具来辅助诊断。如果需要调用工具，请明确要调用的工具名称和输入参数；如果不需要调用工具，请直接给出当前步骤的诊断结论。
`

	StepAnalysisTemplate = `
## 当前任务目标
{{.TaskGoal}}

## 工具执行结果
{{if .ExecutedTools}}
{{range $key, $val := .ExecutedTools}}**{{$key}}**:
{{$val}}

{{end}}
{{else}}
暂无工具执行结果
{{end}}

## 分析要求
请根据上述任务目标和工具执行结果，提取本次诊断任务的关键信息，包括：
1. 发现的关键问题或异常
2. 需要关注的核心指标或信息

请用结构化的方式总结关键信息。
`

	ReportTemplate = `
# 诊断报告

## 执行摘要
本次诊断执行了以下步骤，详细记录如下：

## 执行记录

{{if .ExecutionHistory}}
{{range .ExecutionHistory}}
### {{.Details}}

**执行结果：**
{{.Result}}

---

{{end}}
{{else}}
暂无执行记录
{{end}}

## 总结
基于以上诊断执行记录，本次检查已完成。请根据各步骤的执行结果进行总结分析，内容包括：
1. 诊断结论：总结本次诊断的主要发现和结论。
2. 详细分析：对每个执行步骤的结果进行详细分析，指出发现的问题、异常或需要关注的指标。
3. 后续建议：基于诊断结果，给出后续的建议或行动方案。
`
)
