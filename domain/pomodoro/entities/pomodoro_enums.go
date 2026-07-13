package entities

// PomodoroType 番茄工作类型
type PomodoroType uint8

const (
	PomodoroTypeFocus     PomodoroType = 0
	PomodoroTypeBreak     PomodoroType = 1
	PomodoroTypeLongBreak PomodoroType = 2
)
