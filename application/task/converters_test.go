package task

import (
	"reflect"
	"testing"
	"time"

	"naotodoserver/application/task/dto"
	"naotodoserver/domain/task/entities"
	"naotodoserver/domain/task/valueobjects"
	domaintypes "naotodoserver/domain/types"
)

func TestSplitProjectIds(t *testing.T) {
	const userId int64 = 1001
	tests := []struct {
		name    string
		raw     string
		want    []int64
		wantErr bool
	}{
		{name: "空串 = 无过滤", raw: "", want: nil},
		{name: "单合法", raw: "111", want: []int64{111}},
		{name: "单 inbox", raw: "inbox", want: []int64{userId}},
		{name: "多值保留顺序", raw: "111,222,333", want: []int64{111, 222, 333}},
		{name: "inbox 混入多值", raw: "111,inbox,333", want: []int64{111, userId, 333}},
		{name: "非法段跳过", raw: "111,abc,333", want: []int64{111, 333}},
		{name: "全非法报错", raw: "abc,def", wantErr: true},
		{name: "纯逗号报错", raw: ",,", wantErr: true},
		{name: "去重", raw: "111,111,222", want: []int64{111, 222}},
		{name: "空白段与首尾空格", raw: " 111 , ,222", want: []int64{111, 222}},
		{name: "非正数段跳过", raw: "111,-5,222", want: []int64{111, 222}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := splitProjectIds(tt.raw, userId)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("splitProjectIds(%q) 期望报错，实际 nil", tt.raw)
				}
				return
			}
			if err != nil {
				t.Fatalf("splitProjectIds(%q) 意外错误: %v", tt.raw, err)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("splitProjectIds(%q) = %v, 期望 %v", tt.raw, got, tt.want)
			}
		})
	}
}

func TestSplitTagIds(t *testing.T) {
	tests := []struct {
		name string
		raw  string
		want []string
	}{
		{name: "空串 = 无过滤", raw: "", want: nil},
		{name: "单值", raw: "tag1", want: []string{"tag1"}},
		{name: "多值保留顺序", raw: "tag1,tag2,tag3", want: []string{"tag1", "tag2", "tag3"}},
		{name: "空段去除", raw: "tag1,,tag2", want: []string{"tag1", "tag2"}},
		{name: "去重", raw: "tag1,tag1,tag2", want: []string{"tag1", "tag2"}},
		{name: "纯逗号 = 无过滤", raw: ",,,", want: nil},
		{name: "首尾空格", raw: " tag1 , tag2 ", want: []string{"tag1", "tag2"}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := splitTagIds(tt.raw)
			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("splitTagIds(%q) = %v, 期望 %v", tt.raw, got, tt.want)
			}
		})
	}
}

// TestUpdateTaskReqToValueObject_ProjectId 更新路径 projectId 归一语义（CAL-03 BE-1/2/3）：
// nil（缺省/null，反序列化后不可区分）= 不改；”/inbox = 默认收件箱 userId；合法 ID = 覆盖。
func TestUpdateTaskReqToValueObject_ProjectId(t *testing.T) {
	const userId int64 = 1001
	strPtr := func(s string) *string { return &s }
	int64Ptr := func(v int64) *int64 { return &v }
	tests := []struct {
		name      string
		projectId *string
		want      *int64
	}{
		{name: "缺省/nil = 不改", projectId: nil, want: nil},
		{name: "空串 = 收件箱 userId", projectId: strPtr(""), want: int64Ptr(userId)},
		{name: "inbox 字面量 = 收件箱 userId", projectId: strPtr("inbox"), want: int64Ptr(userId)},
		{name: "合法 ID = 覆盖", projectId: strPtr("222"), want: int64Ptr(222)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			vo, err := UpdateTaskReqToValueObject(userId, &dto.UpdateTaskReq{ProjectId: tt.projectId})
			if err != nil {
				t.Fatalf("UpdateTaskReqToValueObject 意外错误: %v", err)
			}
			if !reflect.DeepEqual(vo.ProjectId, tt.want) {
				t.Fatalf("ProjectId 转换 = %v, 期望 %v", vo.ProjectId, tt.want)
			}
		})
	}
}

// TestCreateTaskReqToValueObject_NullableTimeStartAt SYNC-DEF-01：create/push 请求
// 缺省（nil）与显式空串（清空）必须可区分，且显式空串不被 FillStartAt 复活。
func TestCreateTaskReqToValueObject_NullableTimeStartAt(t *testing.T) {
	strPtr := func(s string) *string { return &s }
	endAt := time.Now().Add(time.Hour).Format(time.RFC3339)

	t.Run("缺省 startAt + endAt 有效 ⇒ 兜底派生", func(t *testing.T) {
		vo, err := CreateTaskReqToValueObject(1001, &dto.CreateTaskReq{
			Name: "t", State: "pending", Priority: "medium", EndAt: strPtr(endAt),
		})
		if err != nil {
			t.Fatalf("意外错误: %v", err)
		}
		if vo.StartAt.IsNull {
			t.Fatal("缺省 startAt 应兜底派生为非空")
		}
	})

	t.Run("显式空串 startAt + endAt 有效 ⇒ 保持清空", func(t *testing.T) {
		vo, err := CreateTaskReqToValueObject(1001, &dto.CreateTaskReq{
			Name: "t", State: "pending", Priority: "medium",
			StartAt: strPtr(""), EndAt: strPtr(endAt),
		})
		if err != nil {
			t.Fatalf("意外错误: %v", err)
		}
		if !vo.StartAt.Valid || !vo.StartAt.IsNull {
			t.Fatalf("显式空串 startAt 应保持清空（Valid=true,IsNull=true），got Valid=%v IsNull=%v Time=%v",
				vo.StartAt.Valid, vo.StartAt.IsNull, vo.StartAt.Time)
		}
	})

	t.Run("有效 startAt ⇒ 指定值", func(t *testing.T) {
		startAt := time.Now().Add(2 * time.Hour).Truncate(time.Second)
		vo, err := CreateTaskReqToValueObject(1001, &dto.CreateTaskReq{
			Name: "t", State: "pending", Priority: "medium",
			StartAt: strPtr(startAt.Format(time.RFC3339)), EndAt: strPtr(endAt),
		})
		if err != nil {
			t.Fatalf("意外错误: %v", err)
		}
		if got, ok := vo.StartAt.Value(); !ok || !got.Equal(startAt) {
			t.Fatalf("有效 startAt 转换错误: got %v ok=%v, want %v", got, ok, startAt)
		}
	})
}

// TestCreateTaskReqToValueObject_StatusTimestamps DEF-SYNC-06：create/push 三状态
// 时间戳字段（archivedAt/starMarkAt/givenUpAt）三态语义必须与可空时间字段一致：
// nil（缺省）⇒ Valid=false 不写列；""（显式清空）⇒ Valid=true,IsNull=true；
// 合法时间 ⇒ 设值；非法非空串 ⇒ 视同缺省（沿用 nullableTimeFromCreateReq 语义）。
func TestCreateTaskReqToValueObject_StatusTimestamps(t *testing.T) {
	strPtr := func(s string) *string { return &s }
	valid := time.Now().Add(2 * time.Hour).Truncate(time.Second)

	cols := []struct {
		name string
		set  func(*dto.CreateTaskReq, *string)
		read func(*valueobjects.CreateTask) domaintypes.NullableTime
	}{
		{"archivedAt", func(r *dto.CreateTaskReq, v *string) { r.ArchivedAt = v }, func(vo *valueobjects.CreateTask) domaintypes.NullableTime { return vo.ArchivedAt }},
		{"starMarkAt", func(r *dto.CreateTaskReq, v *string) { r.StarMarkAt = v }, func(vo *valueobjects.CreateTask) domaintypes.NullableTime { return vo.StarMarkAt }},
		{"givenUpAt", func(r *dto.CreateTaskReq, v *string) { r.GivenUpAt = v }, func(vo *valueobjects.CreateTask) domaintypes.NullableTime { return vo.GivenUpAt }},
	}
	build := func() *dto.CreateTaskReq {
		return &dto.CreateTaskReq{Name: "t", State: "pending", Priority: "medium"}
	}

	for _, c := range cols {
		t.Run(c.name+"/缺省=不写", func(t *testing.T) {
			vo, err := CreateTaskReqToValueObject(1001, build())
			if err != nil {
				t.Fatalf("意外错误: %v", err)
			}
			if nt := c.read(vo); nt.Valid {
				t.Fatalf("%s 缺省时 Valid 应为 false（不写列），got %+v", c.name, nt)
			}
		})
		t.Run(c.name+"/空串=清空", func(t *testing.T) {
			req := build()
			c.set(req, strPtr(""))
			vo, err := CreateTaskReqToValueObject(1001, req)
			if err != nil {
				t.Fatalf("意外错误: %v", err)
			}
			nt := c.read(vo)
			if !nt.Valid || !nt.IsNull {
				t.Fatalf("%s 显式空串应为 Valid=true,IsNull=true（写 NULL），got %+v", c.name, nt)
			}
		})
		t.Run(c.name+"/合法时间=设值", func(t *testing.T) {
			req := build()
			c.set(req, strPtr(valid.Format(time.RFC3339)))
			vo, err := CreateTaskReqToValueObject(1001, req)
			if err != nil {
				t.Fatalf("意外错误: %v", err)
			}
			nt := c.read(vo)
			got, ok := nt.Value()
			if !ok || !got.Equal(valid) {
				t.Fatalf("%s 合法时间转换错误: got %v ok=%v, want %v", c.name, got, ok, valid)
			}
		})
		t.Run(c.name+"/非法非空串=缺省", func(t *testing.T) {
			req := build()
			c.set(req, strPtr("not-a-time"))
			vo, err := CreateTaskReqToValueObject(1001, req)
			if err != nil {
				t.Fatalf("意外错误: %v", err)
			}
			if nt := c.read(vo); nt.Valid {
				t.Fatalf("%s 非法非空串应视同缺省（Valid=false），got %+v", c.name, nt)
			}
		})
	}
}

// TestTaskEntityToGetRes_Counts 领域统计属性：TaskEntityToGetRes 透传计数（ADR §5.1）
func TestTaskEntityToGetRes_Counts(t *testing.T) {
	e := &entities.Task{
		Name:           "t",
		CheckItemCount: 3,
		CommentCount:   5,
		SubtaskCount:   7,
	}
	res := TaskEntityToGetRes(e)
	if res.CheckItemCount != 3 || res.CommentCount != 5 || res.SubtaskCount != 7 {
		t.Fatalf("计数透传错误: got %d/%d/%d, want 3/5/7",
			res.CheckItemCount, res.CommentCount, res.SubtaskCount)
	}
}
