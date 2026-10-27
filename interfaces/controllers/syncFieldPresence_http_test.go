package controllers_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	projectApp "naotodoserver/application/project"
	projectDto "naotodoserver/application/project/dto"
	tagApp "naotodoserver/application/tag"
	tagDto "naotodoserver/application/tag/dto"
	taskApp "naotodoserver/application/task"
	taskDto "naotodoserver/application/task/dto"
	domaintypes "naotodoserver/domain/types"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/controllers"
	"naotodoserver/interfaces/middlewares"

	"github.com/gin-gonic/gin"
)

// B-63 / F-06 载荷存在性回归：控制器必须把「未提供」与「显式置空」区分后传入应用层。
// 这是 update map「未提供不写列」的前置条件（解错则 PR 的修法对线上无效）。

// fakePresenceCheckItemApp 记录 CreateTaskCheckItem 入参
type fakePresenceCheckItemApp struct {
	taskApp.TaskCheckItemApp
	lastReq *taskDto.CreateTaskCheckItemReq
}

func (f *fakePresenceCheckItemApp) CreateTaskCheckItem(
	_ context.Context, _ int64, req *taskDto.CreateTaskCheckItemReq,
) (*taskDto.CreateTaskCheckItemRes, domaintypes.UpsertResult, error) {
	f.lastReq = req
	return &taskDto.CreateTaskCheckItemRes{
		Id: "1", UpdatedAt: "2026-10-27T00:00:00.000Z",
	}, domaintypes.UpsertResult{}, nil
}

// fakePresenceCommentApp 记录 CreateTaskComment 入参
type fakePresenceCommentApp struct {
	taskApp.TaskCommentApp
	lastReq *taskDto.CreateTaskCommentReq
}

func (f *fakePresenceCommentApp) CreateTaskComment(
	_ context.Context, _ int64, req *taskDto.CreateTaskCommentReq,
) (*taskDto.TaskCommentRes, domaintypes.UpsertResult, error) {
	f.lastReq = req
	return &taskDto.TaskCommentRes{
		Id: "1", UpdatedAt: "2026-10-27T00:00:00.000Z",
	}, domaintypes.UpsertResult{}, nil
}

// fakePresenceProjectApp 记录 Create 入参
type fakePresenceProjectApp struct {
	projectApp.ProjectApp
	lastReq *projectDto.CreateProjectReq
}

func (f *fakePresenceProjectApp) Create(
	_ context.Context, _ int64, req *projectDto.CreateProjectReq,
) (*projectDto.CreateProjectRes, domaintypes.UpsertResult, error) {
	f.lastReq = req
	return &projectDto.CreateProjectRes{
		Id: "1", UpdatedAt: "2026-10-27T00:00:00.000Z",
	}, domaintypes.UpsertResult{}, nil
}

// fakePresenceTagApp 记录 CreateTag 入参
type fakePresenceTagApp struct {
	tagApp.TagApp
	lastReq *tagDto.CreateTagReq
}

func (f *fakePresenceTagApp) CreateTag(
	_ context.Context, _ int64, req *tagDto.CreateTagReq,
) (*tagDto.CreateTagRes, domaintypes.UpsertResult, error) {
	f.lastReq = req
	return &tagDto.CreateTagRes{
		Id: "1", UpdatedAt: "2026-10-27T00:00:00.000Z",
	}, domaintypes.UpsertResult{}, nil
}

type presenceFakes struct {
	checkItem *fakePresenceCheckItemApp
	comment   *fakePresenceCommentApp
	project   *fakePresenceProjectApp
	tag       *fakePresenceTagApp
}

// newPresenceEngine 挂载真实 SyncController（其余表用 nil/fake 兜底）
func newPresenceEngine(fakes *presenceFakes) *gin.Engine {
	gin.SetMode(gin.TestMode)
	engine := gin.New()
	engine.Use(middlewares.BodyLimit(2<<20, nil))
	ctrl := controllers.NewSyncController(
		nil, fakes.checkItem, fakes.comment, fakes.project, fakes.tag, nil,
	)
	engine.POST("/api/sync/push", func(ctx *gin.Context) {
		ctx.Request = ctx.Request.WithContext(iCtx.SetUserId(ctx.Request.Context(), 42))
		ctrl.Push(ctx)
	})
	return engine
}

// postSyncPushRaw 发送原始 JSON（精确控制「键是否存在」）；非 200 直接失败
func postSyncPushRaw(t *testing.T, engine *gin.Engine, body string) {
	t.Helper()
	req := httptest.NewRequest(
		http.MethodPost, "/api/sync/push", strings.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	engine.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("sync push status = %d（body=%s）", rec.Code, rec.Body.String())
	}
}

// TestSyncPushCheckItemDescriptionPresence 检查项 description：缺省 ⇒ nil；值/""/null ⇒ 指针
func TestSyncPushCheckItemDescriptionPresence(t *testing.T) {
	cases := []struct {
		name    string
		body    string
		wantNil bool
		want    string
	}{
		{"未提供", `{"taskCheckItems":[{"taskId":"1","name":"n"}]}`, true, ""},
		{"值", `{"taskCheckItems":[{"taskId":"1","name":"n","description":"备注B"}]}`, false, "备注B"},
		{"显式空串", `{"taskCheckItems":[{"taskId":"1","name":"n","description":""}]}`, false, ""},
		{"显式 null", `{"taskCheckItems":[{"taskId":"1","name":"n","description":null}]}`, false, ""},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fakes := &presenceFakes{checkItem: &fakePresenceCheckItemApp{}}
			postSyncPushRaw(t, newPresenceEngine(fakes), c.body)
			got := fakes.checkItem.lastReq
			if got == nil {
				t.Fatal("应用层未收到检查项入参")
			}
			if c.wantNil {
				if got.Description != nil {
					t.Fatalf("未提供时应为 nil（不写列），实际 %q", *got.Description)
				}
				return
			}
			if got.Description == nil {
				t.Fatal("显式提供时应为非 nil（含清空），实际 nil")
			}
			if *got.Description != c.want {
				t.Fatalf("Description = %q, want %q", *got.Description, c.want)
			}
		})
	}
}

// TestSyncPushCommentPresence 评论 attachments / isTopUp：缺省 ⇒ nil；显式 ⇒ 指针
func TestSyncPushCommentPresence(t *testing.T) {
	t.Run("未提供", func(t *testing.T) {
		fakes := &presenceFakes{comment: &fakePresenceCommentApp{}}
		postSyncPushRaw(t, newPresenceEngine(fakes),
			`{"taskComments":[{"taskId":"1","content":"c"}]}`)
		got := fakes.comment.lastReq
		if got.Attachments != nil || got.IsTopUp != nil {
			t.Fatalf("缺省应为 nil，实际 attachments=%v isTopUp=%v", got.Attachments, got.IsTopUp)
		}
	})
	t.Run("显式提供", func(t *testing.T) {
		fakes := &presenceFakes{comment: &fakePresenceCommentApp{}}
		postSyncPushRaw(t, newPresenceEngine(fakes),
			`{"taskComments":[{"taskId":"1","content":"c",`+
				`"attachments":["f1"],"isTopUp":true}]}`)
		got := fakes.comment.lastReq
		if got.Attachments == nil || len(*got.Attachments) != 1 || (*got.Attachments)[0] != "f1" {
			t.Fatalf("attachments 未透传: %v", got.Attachments)
		}
		if got.IsTopUp == nil || !*got.IsTopUp {
			t.Fatalf("isTopUp 未透传: %v", got.IsTopUp)
		}
	})
}

// TestSyncPushProjectTagSortIdPresence sortId：缺省 ⇒ nil（不写列）；显式 0/非 0 ⇒ 指针
func TestSyncPushProjectTagSortIdPresence(t *testing.T) {
	cases := []struct {
		name    string
		kind    string
		body    string
		wantNil bool
		want    uint16
	}{
		{"project 未提供", "project", `{"projects":[{"name":"p"}]}`, true, 0},
		{"project 显式 0", "project", `{"projects":[{"name":"p","sortId":0}]}`, false, 0},
		{"project 非 0", "project", `{"projects":[{"name":"p","sortId":7}]}`, false, 7},
		{"tag 未提供", "tag", `{"tags":[{"name":"t","color":"#fff"}]}`, true, 0},
		{"tag 显式 0", "tag", `{"tags":[{"name":"t","color":"#fff","sortId":0}]}`, false, 0},
		{"tag 非 0", "tag", `{"tags":[{"name":"t","color":"#fff","sortId":9}]}`, false, 9},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			fakes := &presenceFakes{
				project: &fakePresenceProjectApp{},
				tag:     &fakePresenceTagApp{},
			}
			postSyncPushRaw(t, newPresenceEngine(fakes), c.body)

			var got *uint16
			if c.kind == "project" {
				got = fakes.project.lastReq.SortId
			} else {
				got = fakes.tag.lastReq.SortId
			}
			if c.wantNil {
				if got != nil {
					t.Fatalf("未提供时应为 nil（不写列），实际 %d", *got)
				}
				return
			}
			if got == nil || *got != c.want {
				t.Fatalf("SortId 透传错误: %v, want %d", got, c.want)
			}
		})
	}
}
