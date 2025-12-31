package consts

var TodoStateMap = map[string]int8{
	"todo":        1,
	"in-progress": 2,
	"done":        3,
}

var TodoStateMapReverse = map[int8]string{
	1: "todo",
	2: "in-progress",
	3: "done",
}

var TodoPriorityMap = map[string]int8{
	"low":    1,
	"medium": 2,
	"high":   3,
	"urgent": 4,
}

var TodoPriorityMapReverse = map[int8]string{
	1: "low",
	2: "medium",
	3: "high",
	4: "urgent",
}
