package prompt

const (
	Middleware       = "Middleware"
	ExecutionHistory = "ExecutionHistory"
	TaskGoal         = "TaskGoal"
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
当前需要执行任务: {{.TaskGoal}}
`
)
