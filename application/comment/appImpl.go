package comment

import (
	"context"
	"errors"
	"naotodoserver/domain/comment/service"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"
	"strconv"
)

func RegistDomainImpl(commentDomain service.CommentDomain) CommentApp {
	once.Do(func() {
		App = &CommentAppImpl{
			CommentDomain: commentDomain,
		}
	})
	return App
}

// GetComment 获取评论详情
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
	return CommentEntity2Res(commentEntity), nil
}

// CreateComment 创建评论
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
	commentEntity := CreateCommentReq2Entity(req)
	// 3. 创建评论
	commentEntity, err := commentApp.CommentDomain.Create(ctx, userId, commentEntity)
	if err != nil {
		return nil, err
	}
	// 4. 返回评论详情
	return CommentEntity2Res(commentEntity), nil
}

// UpdateComment 更新评论
func (commentApp *CommentAppImpl) UpdateComment(
	ctx context.Context,
	commentId string,
	req *types.UpdateCommentReq,
) (*types.UpdateCommentRes, error) {
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
	// 3. 转换请求体为实体
	updateEntity := UpdateCommentReq2Entity(req)
	// 4. 更新评论
	err = commentApp.CommentDomain.Update(ctx, userId, commentId64, updateEntity)
	if err != nil {
		return nil, err
	}
	// 5. 返回评论 ID
	return &types.UpdateCommentRes{CommentId: commentId}, nil
}

// DeleteComment 删除评论
func (commentApp *CommentAppImpl) DeleteComment(
	ctx context.Context,
	commentId string,
) (*types.DeleteCommentRes, error) {
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
	// 3. 删除评论
	err = commentApp.CommentDomain.Delete(ctx, userId, commentId64)
	if err != nil {
		return nil, err
	}
	// 4. 返回评论 ID
	return &types.DeleteCommentRes{CommentId: commentId}, nil
}

// ListComment 获取评论列表
func (commentApp *CommentAppImpl) ListComment(
	ctx context.Context,
	taskId string,
) (types.ListCommentRes, error) {
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
	commentEntityList, err := commentApp.CommentDomain.List(ctx, userId, taskId64)
	if err != nil {
		return nil, err
	}
	// 4. 返回评论列表
	return CommentEntities2ResList(commentEntityList), nil
}
