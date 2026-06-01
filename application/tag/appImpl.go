package tag

import (
	"context"
	"errors"
	"naotodoserver/domain/tag/service"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"
	"strconv"
)

// NewTagApp 创建标签应用层实例
func NewTagApp(tagDomain service.TagDomain) TagApp {
	impl := &TagAppImpl{tagDomain: tagDomain}
	return impl
}

// 获取标签信息
// @param ctx 上下文
// @param tagId 标签 ID
// @return 标签响应体
// @return error
func (tagApp *TagAppImpl) GetTag(
	ctx context.Context,
	tagId string,
) (*types.GetTagRes, error) {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 转换 tagId
	tagId64, err := strconv.ParseInt(tagId, 10, 64)
	if err != nil {
		return nil, errors.New("标签 ID 格式错误")
	}
	// 获取标签信息
	tagEntity, err := tagApp.tagDomain.GetById(ctx, userId, tagId64)
	if err != nil {
		return nil, err
	}
	// 实体转换响应体
	getRes := TagEntityToGetRes(tagEntity)
	// 返回结果
	return getRes, nil
}

// 创建标签
// @param ctx 上下文
// @param tagId 标签 ID
// @param req 创建标签请求体
// @return 创建标签响应体
// @return error
func (tagApp *TagAppImpl) CreateTag(
	ctx context.Context,
	createTagReq *types.CreateTagReq,
) (*types.CreateTagRes, error) {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 转换请求体
	createTagValueObject, err := CreateTagReqToValueObject(createTagReq)
	if err != nil {
		return nil, err
	}
	// 创建标签
	tagEntity, err := tagApp.tagDomain.Create(ctx, userId, createTagValueObject)
	if err != nil {
		return nil, err
	}
	// 实体转换响应体并返回结果
	return TagEntityToCreateRes(tagEntity), nil
}

// 更新标签信息
// @param ctx 上下文
// @param tagId 标签 ID
// @param updateTagReq 更新标签请求体
// @return error
func (tagApp *TagAppImpl) UpdateTag(
	ctx context.Context,
	tagId string,
	updateTagReq *types.UpdateTagReq,
) error {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	// 转换 tagId
	tagId64, err := strconv.ParseInt(tagId, 10, 64)
	if err != nil {
		return errors.New("标签 ID 格式错误")
	}
	// 转换请求体
	updateTagValueObject, err := UpdateTagReqToValueObject(updateTagReq)
	if err != nil {
		return err
	}
	// 更新标签信息
	err = tagApp.tagDomain.Update(ctx, userId, tagId64, updateTagValueObject)
	if err != nil {
		return err
	}
	// 转换结果并返回
	return nil
}

// 删除标签
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
		return errors.New("用户 ID 无效")
	}
	// 转换 tagId
	tagId64, err := strconv.ParseInt(tagId, 10, 64)
	if err != nil {
		return errors.New("标签 ID 格式错误")
	}
	// 删除标签信息
	err = tagApp.tagDomain.Delete(ctx, userId, tagId64)
	if err != nil {
		return err
	}
	// 转换结果并返回
	return nil
}

// 获取所有标签信息
// @param ctx 上下文
// @return 标签响应体
// @return error
func (tagApp *TagAppImpl) ListTag(
	ctx context.Context,
) ([]*types.GetTagRes, error) {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 获取所有标签信息
	tagEntities, err := tagApp.tagDomain.List(ctx, userId)
	if err != nil {
		return nil, err
	}
	// 转换结果并返回
	return TagEntitiesToGetResList(tagEntities), nil
}

// 批量更新标签
func (tagApp *TagAppImpl) BatchUpdateTags(
	ctx context.Context,
	req *types.BatchUpdateTagReq,
) (*types.BatchUpdateTagRes, error) {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 请求体转换值对象
	batchVOs, err := BatchUpdateTagReqToValueObjects(req)
	if err != nil {
		return nil, err
	}
	// 调用域函数 - 批量更新标签
	updatedEntities, err := tagApp.tagDomain.BatchUpdate(ctx, userId, batchVOs)
	if err != nil {
		return nil, err
	}
	// 实体转换响应体
	tagResList := TagEntitiesToGetResList(updatedEntities)
	// 返回结果
	return &types.BatchUpdateTagRes{
		UpdatedCount: int64(len(tagResList)),
		Tags:         tagResList,
	}, nil
}

// 获取标签偏好设置
// @param ctx 上下文
// @param tagId 标签 ID
// @return 标签偏好设置响应体
// @return error
func (tagApp *TagAppImpl) GetTagPreference(
	ctx context.Context,
	tagId string,
) (*types.GetTagPreferenceRes, error) {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 获取标签 ID
	tagId64, err := strconv.ParseInt(tagId, 10, 64)
	if err != nil {
		return nil, errors.New("标签 ID 格式错误")
	}
	// 获取标签偏好设置
	tagPreferenceEntity, err := tagApp.tagDomain.GetPreference(ctx, userId, tagId64)
	if err != nil {
		return nil, err
	}
	// 响应体转换并返回结果
	return TagPreferenceEntityToGetRes(tagPreferenceEntity), nil

}

// 更新标签偏好设置
// @param ctx 上下文
// @param tagId 标签 ID
// @param req 更新标签偏好设置请求体
// @return error
func (tagApp *TagAppImpl) UpdateTagPreference(
	ctx context.Context,
	tagId string,
	updateTagPreferenceReq *types.UpdateTagPreferenceReq,
) error {
	// 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return errors.New("用户 ID 无效")
	}
	// 获取标签 ID
	tagId64, err := strconv.ParseInt(tagId, 10, 64)
	if err != nil {
		return errors.New("标签 ID 格式错误")
	}
	// 数据转换
	saveTagPreferenceValueObject, err := UpdateTagPreferenceReqToValueObject(updateTagPreferenceReq)
	if err != nil {
		return err
	}
	// 更新标签偏好设置
	err = tagApp.tagDomain.UpdatePreference(
		ctx,
		userId,
		tagId64,
		saveTagPreferenceValueObject,
	)
	if err != nil {
		return err
	}
	// 返回结果
	return nil
}
