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

// RemindRepeatMap 提醒重复类型映射
var RemindRepeatMap = map[string]int8{
	"none":    0,
	"daily":   1,
	"weekly":  2,
	"monthly": 3,
}

// RemindRepeatMapReverse 提醒重复类型反向映射
var RemindRepeatMapReverse = map[int8]string{
	0: "none",
	1: "daily",
	2: "weekly",
	3: "monthly",
}

// WeekdayBitmask 星期到位掩码的映射
var WeekdayBitmask = map[int]int8{
	0: 1,  // 周日
	1: 2,  // 周一
	2: 4,  // 周二
	3: 8,  // 周三
	4: 16, // 周四
	5: 32, // 周五
	6: 64, // 周六
}

// WeekdayBitmaskReverse 位掩码到星期的反向映射
var WeekdayBitmaskReverse = map[int8]int{
	1: 0,  // 周日
	2: 1,  // 周一
	4: 2,  // 周二
	8: 3,  // 周三
	16: 4, // 周四
	32: 5, // 周五
	64: 6, // 周六
}
