package apis

import (
	"errors"
	"fmt"
	"naotodoserver/models"
	"naotodoserver/utils"
	"strconv"
	"strings"
	"time"

	"gorm.io/gorm"
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

func ParseRelativeDateToUpdateCond(tx *gorm.DB, relativeDate string) {
	fmt.Println("relativeDate: ", relativeDate)
	switch relativeDate {
	case "today":
		tx.Where("end_at >= ?", time.Now().Format("2006-01-02"))
	case "tomorrow":
		tx.Where("end_at >= ?", time.Now().AddDate(0, 0, 1).Format("2006-01-02"))
	case "week":
		{
			start, end := utils.GetWeekRange(time.Now())
			tx.Where("end_at >= ? and end_at <= ?", start, end)
		}
	case "month":
		tx.Where("end_at >= ?", time.Now().AddDate(0, 0, 7).Format("2006-01-02"))
	case "-today":
		tx.Where("end_at < ?", time.Now().Format("2006-01-02"))
	}
}

func ParseSortStringToQueryCond(tx *gorm.DB, sort string) {
	var splited = strings.Split(sort, ":")
	if len(splited) != 2 {
		return
	}
	tx.Order(utils.ToSnakeCase(splited[0]) + " " + splited[1])
}
