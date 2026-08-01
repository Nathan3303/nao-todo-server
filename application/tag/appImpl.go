package tag

import (
	"context"
	"fmt"
	"naotodoserver/application/idutil"
	"naotodoserver/application/tag/dto"
	domerr "naotodoserver/domain/errors"
	"naotodoserver/domain/tag/repositories"
	"naotodoserver/domain/tag/service"
	iCtx "naotodoserver/infrastructure/context"
)

// NewTagApp 创建标签应用层实例
func NewTagApp(
	tagDomain service.TagDomain,
	tagRepo repositories.TagRepository,
	preferenceRepo repositories.TagPreference,
) TagApp {
	impl := &TagAppImpl{
		tagDomain:      tagDomain,
		tagRepo:        tagRepo,
		preferenceRepo: preferenceRepo,
	}
	return impl
}

// GetTag 获取标签信息
// @param ctx 上下文
// @param tagId 标签 ID
// @return 标签响应体
// @return error
func (tagApp *TagAppImpl) GetTag(
	ctx context.Context,
	tagId string,
) (*dto.GetTagRes, error) {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	// 转换 tagId
	tagId64, err := idutil.ParseID(tagId)
	if err != nil {
		return nil, domerr.ErrInvalidTagID
	}
	// 获取标签信息
	tagEntity, err := tagApp.tagRepo.GetById(ctx, userId, tagId64)
	if err != nil {
		return nil, fmt.Errorf("tag.Get: %w", err)
	}
	// 实体转换响应体
	getRes := TagEntityToGetRes(tagEntity)
	// 返回结果
	return getRes, nil
}

// CreateTag 创建标签
// @param ctx 上下文
// @param tagId 标签 ID
// @param req 创建标签请求体
// @return 创建标签响应体
// @return error
func (tagApp *TagAppImpl) CreateTag(
	ctx context.Context,
	createTagReq *dto.CreateTagReq,
) (*dto.CreateTagRes, error) {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	// 转换请求体
	createTagValueObject, err := CreateTagReqToValueObject(createTagReq)
	if err != nil {
		return nil, err
	}
	// 创建标签
	tagEntity, err := tagApp.tagDomain.Create(ctx, userId, createTagValueObject)
	if err != nil {
		return nil, fmt.Errorf("tag.Create: %w", err)
	}
	// 实体转换响应体并返回结果
	return TagEntityToCreateRes(tagEntity), nil
}

// UpdateTag 更新标签信息
// @param ctx 上下文
// @param tagId 标签 ID
// @param updateTagReq 更新标签请求体
// @return error
func (tagApp *TagAppImpl) UpdateTag(
	ctx context.Context,
	tagId string,
	updateTagReq *dto.UpdateTagReq,
) error {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return domerr.ErrInvalidUserID
	}
	// 转换 tagId
	tagId64, err := idutil.ParseID(tagId)
	if err != nil {
		return domerr.ErrInvalidTagID
	}
	// 转换请求体
	updateTagValueObject, err := UpdateTagReqToValueObject(updateTagReq)
	if err != nil {
		return err
	}
	// 更新标签信息
	if err := tagApp.tagRepo.Update(ctx, userId, tagId64, updateTagValueObject); err != nil {
		return fmt.Errorf("tag.Update: %w", err)
	}
	return nil
}

// DeleteTag 删除标签
// @param ctx 上下文
// @param tagId 标签 ID
// @return error
func (tagApp *TagAppImpl) DeleteTag(
	ctx context.Context,
	tagId string,
) error {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return domerr.ErrInvalidUserID
	}
	// 转换 tagId
	tagId64, err := idutil.ParseID(tagId)
	if err != nil {
		return domerr.ErrInvalidTagID
	}
	// 删除标签信息
	err = tagApp.tagDomain.Delete(ctx, userId, tagId64)
	if err != nil {
		return fmt.Errorf("tag.Delete: %w", err)
	}
	// 转换结果并返回
	return nil
}

// ListTag 获取所有标签信息
// @param ctx 上下文
// @return 标签响应体
// @return error
func (tagApp *TagAppImpl) ListTag(
	ctx context.Context,
) ([]*dto.GetTagRes, error) {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	// 获取所有标签信息
	tagEntities, err := tagApp.tagRepo.Get(ctx, userId)
	if err != nil {
		return nil, err
	}
	// 转换结果并返回
	return TagEntitiesToGetResList(tagEntities), nil
}

// ListTagByIds 根据标签ID列表获取标签列表
// @param ctx 上下文
// @param tagIds 标签 ID列表
// @return 标签响应体列表
// @return error
func (tagApp *TagAppImpl) ListTagByIds(
	ctx context.Context,
	tagIds []string,
) ([]*dto.GetTagRes, error) {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	// 转换 tagIds 为 int64 列表 - idutil.ParseID(tagId)
	var tagIds64 []int64
	for _, tagId := range tagIds {
		tagId64, err := idutil.ParseID(tagId)
		if err != nil {
			return nil, domerr.ErrInvalidTagID
		}
		tagIds64 = append(tagIds64, tagId64)
	}
	// 获取标签信息
	tagEntities, err := tagApp.tagRepo.GetByIds(ctx, userId, tagIds64)
	if err != nil {
		return nil, err
	}
	// 转换结果并返回
	return TagEntitiesToGetResList(tagEntities), nil
}

// BatchUpdateTags 批量更新标签
// @param ctx 上下文
// @param req 批量更新标签请求体
// @return 批量更新标签响应体
// @return error
func (tagApp *TagAppImpl) BatchUpdateTags(
	ctx context.Context,
	req *dto.BatchUpdateTagReq,
) (*dto.BatchUpdateTagRes, error) {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	// 请求体转换值对象
	batchVOs, err := BatchUpdateTagReqToValueObjects(req)
	if err != nil {
		return nil, err
	}
	updatedEntities, err := tagApp.tagRepo.BatchUpdate(ctx, userId, batchVOs)
	if err != nil {
		return nil, err
	}
	// 实体转换响应体
	tagResList := TagEntitiesToGetResList(updatedEntities)
	// 返回结果
	return &dto.BatchUpdateTagRes{
		UpdatedCount: int64(len(tagResList)),
		Tags:         tagResList,
	}, nil
}

// GetTagPreference 获取标签偏好设置
// @param ctx 上下文
// @param tagId 标签 ID
// @return 标签偏好设置响应体
// @return error
func (tagApp *TagAppImpl) GetTagPreference(
	ctx context.Context,
	tagId string,
) (*dto.GetTagPreferenceRes, error) {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, domerr.ErrInvalidUserID
	}
	// 获取标签 ID
	tagId64, err := idutil.ParseID(tagId)
	if err != nil {
		return nil, domerr.ErrInvalidTagID
	}
	// 获取标签偏好设置
	tagPreferenceEntity, err := tagApp.preferenceRepo.Get(ctx, userId, tagId64)
	if err != nil {
		return nil, err
	}
	// 响应体转换并返回结果
	return TagPreferenceEntityToGetRes(tagPreferenceEntity), nil
}

// UpdateTagPreference 更新标签偏好设置
// @param ctx 上下文
// @param tagId 标签 ID
// @param req 更新标签偏好设置请求体
// @return error
func (tagApp *TagAppImpl) UpdateTagPreference(
	ctx context.Context,
	tagId string,
	updateTagPreferenceReq *dto.UpdateTagPreferenceReq,
) error {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return domerr.ErrInvalidUserID
	}
	// 获取标签 ID
	tagId64, err := idutil.ParseID(tagId)
	if err != nil {
		return domerr.ErrInvalidTagID
	}
	// 数据转换
	saveTagPreferenceValueObject, err := UpdateTagPreferenceReqToValueObject(updateTagPreferenceReq)
	if err != nil {
		return err
	}
	// 更新标签偏好设置
	return tagApp.preferenceRepo.Save(
		ctx,
		userId,
		tagId64,
		saveTagPreferenceValueObject,
	)
}
