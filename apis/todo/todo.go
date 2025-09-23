package apis

import (
	"errors"
	"fmt"
	"naotodoserver/models"
	"strconv"
	"time"
)

var TodoStateMap = map[string]int8{
	"todo":        0,
	"doing":       1,
	"in-progress": 1,
	"done":        2,
}

var TodoStateMapReverse = map[int8]string{
	0: "todo",
	1: "in-progress",
	2: "done",
}

var TodoPriorityMap = map[string]int8{
	"low":    0,
	"medium": 1,
	"high":   2,
	"urgent": 3,
}

var TodoPriorityMapReverse = map[int8]string{
	0: "low",
	1: "medium",
	2: "high",
	3: "urgent",
}

func ParseDateString(formatString, dateString string) (*time.Time, error) {
	if dateString == "" {
		return nil, errors.New("date is empty")
	}
	fmt.Println(dateString)
	date, err := time.Parse(formatString, dateString)
	if err != nil {
		return nil, err
	}
	return &date, nil
}

func ToTodoResponse(todo *models.Todo) models.TodoResponse {
	var todoResponse models.TodoResponse
	todoResponse.ID = strconv.FormatInt(todo.ID, 10)
	todoResponse.CreatedAt = todo.CreatedAt
	todoResponse.UpdatedAt = todo.UpdatedAt
	todoResponse.DeletedAt = todo.DeletedAt

	todoResponse.UserId = strconv.FormatInt(todo.UserId, 10)
	todoResponse.ProjectId = strconv.FormatInt(todo.ProjectId, 10)
	if todo.ParentTodoId != nil {
		*todoResponse.ParentTodoId = strconv.FormatInt(*todo.ParentTodoId, 10)
	}

	todoResponse.Name = todo.Name
	todoResponse.Description = todo.Description
	todoResponse.State = TodoStateMapReverse[todo.State]
	todoResponse.Priority = TodoPriorityMapReverse[todo.Priority]
	todoResponse.StartAt = todo.StartAt
	todoResponse.EndAt = todo.EndAt
	todoResponse.ArchivedAt = todo.ArchivedAt
	todoResponse.FavoritedAt = todo.FavoritedAt
	todoResponse.GivenUpAt = todo.GivenUpAt
	todoResponse.Tags = todo.Tags

	todoResponse.IsArchived = todo.ArchivedAt != nil
	todoResponse.IsDeleted = todo.DeletedAt.Valid
	todoResponse.IsFavorited = todo.FavoritedAt != nil
	todoResponse.IsGivenUp = todo.GivenUpAt != nil

	return todoResponse
}
