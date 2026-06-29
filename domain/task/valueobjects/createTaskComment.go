package valueobjects

import (
	"errors"

	"naotodoserver/domain/textutils"
)

// CreateTaskComment 创建待办任务评论 Value Object
// 评论内容最多 1000 个字符
// 最多拥有 8 个附件
type CreateTaskComment struct {
	UserId      int64
	TaskId      int64
	Content     string
	Attachments []string
	IsTopUp     bool
}

// Validate 验证创建待办任务评论 Value Object 是否符合要求
// @return error 验证失败时返回错误
func (createComment *CreateTaskComment) Validate() error {
	if createComment.UserId <= 0 {
		return errors.New("用户 Id 不能为空")
	}
	if createComment.TaskId <= 0 {
		return errors.New("待办任务 Id 不能为空")
	}
	if createComment.Content == "" {
		return errors.New("待办任务评论内容不能为空")
	}
	if textutils.RuneLength(createComment.Content) > 1000 {
		return errors.New("待办任务评论内容最多 1000 个字符")
	}
	if len(createComment.Attachments) > 8 {
		return errors.New("最多拥有 8 个附件")
	}
	return nil
}

// NewCreateTaskComment 创建待办任务评论 Value Object
// @param taskId 待办任务 Id
// @param content 评论内容
// @param attachments 评论附件
// @param isTopUp 是否为点赞评论
// @return *CreateTaskComment 创建评论 Value Object
// @return error 创建失败时返回错误
func NewCreateTaskComment(
	userId int64,
	taskId int64,
	content string,
	attachments []string,
	isTopUp bool,
) (*CreateTaskComment, error) {
	vo := &CreateTaskComment{
		UserId:      userId,
		TaskId:      taskId,
		Content:     content,
		Attachments: attachments,
		IsTopUp:     isTopUp,
	}
	err := vo.Validate()
	if err != nil {
		return nil, err
	}
	return vo, nil
}
