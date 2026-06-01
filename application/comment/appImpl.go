package comment

import (
	"context"
	"errors"
	"naotodoserver/domain/comment/service"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"
	"strconv"
)

func NewCommentApp(commentDomain service.CommentDomain) CommentApp {
	impl := &CommentAppImpl{
		CommentDomain: commentDomain,
	}
	return impl
}

// GetComment 获取评论详情
// @param ctx 上下文
// @param commentId 评论 ID
// @return 评论详情
// @return error 错误
func (commentApp *CommentAppImpl) GetComment(
	ctx context.Context,
	commentId string,
) (*types.CommentRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换评论 ID
	commentId64, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		return nil, errors.New("评论 ID 无效")
	}
	// 3. 获取评论详情
	commentEntity, err := commentApp.CommentDomain.GetById(ctx, userId, commentId64)
	if err != nil {
		return nil, err
	}
	// 4. 返回评论详情
	return CommentEntityToRes(commentEntity), nil
}

// CreateComment 创建评论
// @param ctx 上下文
// @param req 创建评论请求体
// @return 评论详情
// @return error 错误信息
func (commentApp *CommentAppImpl) CreateComment(
	ctx context.Context,
	req *types.CreateCommentReq,
) (*types.CommentRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换请求体为实体
	createCommentValueObject, err := CreateCommentReqToValueObject(userId, req)
	if err != nil {
		return nil, err
	}
	// 3. 创建评论
	commentEntity, err := commentApp.CommentDomain.Create(ctx, userId, createCommentValueObject)
	if err != nil {
		return nil, err
	}
	// 4. 返回评论详情
	return CommentEntityToRes(commentEntity), nil
}

// UpdateComment 更新评论
// @param ctx 上下文
// @param commentId 评论 ID
// @param req 更新评论请求体
// @return 评论详情
// @return error 错误
func (commentApp *CommentAppImpl) UpdateComment(
	ctx context.Context,
	commentId string,
	req *types.UpdateCommentReq,
) error {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	// 2. 转换评论 ID
	commentId64, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		return errors.New("评论 ID 无效")
	}
	// 3. 转换请求体为实体
	updateCommentValueObject, err := UpdateCommentReqToValueObject(req)
	if err != nil {
		return err
	}
	// 4. 更新评论
	return commentApp.CommentDomain.Update(
		ctx,
		userId,
		commentId64,
		updateCommentValueObject,
	)
}

// DeleteComment 删除评论
func (commentApp *CommentAppImpl) DeleteComment(
	ctx context.Context,
	commentId string,
) error {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	// 2. 转换评论 ID
	commentId64, err := strconv.ParseInt(commentId, 10, 64)
	if err != nil {
		return errors.New("评论 ID 无效")
	}
	// 3. 删除评论
	return commentApp.CommentDomain.Delete(ctx, userId, commentId64)
}

// ListComment 获取评论列表
func (commentApp *CommentAppImpl) ListComment(
	ctx context.Context,
	taskId string,
) ([]*types.CommentRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换待办任务 ID
	taskId64, err := strconv.ParseInt(taskId, 10, 64)
	if err != nil {
		return nil, errors.New("待办任务 ID 无效")
	}
	// 3. 获取评论列表
	commentEntities, err := commentApp.CommentDomain.List(ctx, userId, taskId64)
	if err != nil {
		return nil, err
	}
	// 4. 返回评论列表
	commentListRes := CommentEntitiesToListRes(commentEntities)
	return commentListRes, nil
}
