package project

import (
	"testing"

	"naotodoserver/domain/project/entities"
)

// TestProjectEntityToGetRes_TaskCount 领域统计属性：ProjectEntityToGetRes 透传 taskCount（ADR §5.1）
func TestProjectEntityToGetRes_TaskCount(t *testing.T) {
	e := &entities.Project{
		Name:      "p",
		TaskCount: 9,
	}
	res := ProjectEntityToGetRes(e)
	if res.TaskCount != 9 {
		t.Fatalf("TaskCount 透传错误: got %d, want 9", res.TaskCount)
	}
}
