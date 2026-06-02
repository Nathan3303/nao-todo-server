package valueobjects

import (
	"errors"

	"naotodoserver/domain/textutils"
)

// UpdateComment 更新评论 Value Object
type UpdateComment struct {
	Content     *string
	Attachments *[]string
	IsTopUp     *bool
}

// Validate 验证创建评论 Value Object 是否符合要求
// @return error 验证失败时返回错误
func (updateComment *UpdateComment) Validate() error {
	if updateComment.Content != nil && textutils.RuneLength(*updateComment.Content) == 0 {
		return errors.New("评论内容不能为空")
	}
	if updateComment.Content != nil && textutils.RuneLength(*updateComment.Content) > 512 {
		return errors.New("评论内容最多 512 个字符")
	}
	if updateComment.Attachments != nil && len(*updateComment.Attachments) > 8 {
		return errors.New("最多拥有 8 个附件")
	}
	return nil
}

// NewUpdateComment 更新评论 Value Object
// @param content 评论内容
// @param attachments 评论附件
// @param isTopUp 是否为点赞评论
// @return *UpdateComment 更新评论 Value Object
// @return error 更新失败时返回错误
func NewUpdateComment(
	content *string,
	attachments *[]string,
	isTopUp *bool,
) (*UpdateComment, error) {
	vo := &UpdateComment{
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
