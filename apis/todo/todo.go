package apis

import (
	"errors"
	"time"
)

var TodoStateMap = map[string]int8{
	"todo":        0,
	"doing":       1,
	"in-progress": 1,
	"done":        2,
}

var TodoPriorityMap = map[string]int8{
	"low":    0,
	"medium": 1,
	"high":   2,
	"urgent": 3,
}

func GetTodoState(stateString string) int8 {
	state := TodoStateMap[stateString]
	return state
}

func GetTodoPriority(priorityString string) int8 {
	priority := TodoPriorityMap[priorityString]
	return priority
}

func ParseDateString(formatString, dateString string) (*time.Time, error) {
	if dateString == "" {
		return nil, errors.New("date is empty")
	}
	date, err := time.Parse(formatString, dateString)
	if err != nil {
		return nil, err
	}
	return &date, nil
}
