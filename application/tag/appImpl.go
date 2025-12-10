package tag

import (
	"context"
	"errors"
	"naotodoserver/domain/tag/service"
	iCtx "naotodoserver/infrastructure/context"
	"naotodoserver/interfaces/types"
	"strconv"
)

// 注册函数
func RegistDomainImpl(tagDomain service.TagDomain) TagApp {
	once.Do(func() {
		App.tagDomain = tagDomain
	})
	return &App
}

/*
 * Get tag
 * 获取单个标签信息
 */
func (tagApp *TagAppImpl) GetTag(
	ctx context.Context,
	tagId string,
) (*types.GetTagRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换 tagId
	tagId64, err := strconv.ParseInt(tagId, 10, 64)
	if err != nil {
		return nil, errors.New("标签 ID 格式错误")
	}
	// 2. 调用域函数 - 获取标签信息
	tagEntity, err := tagApp.tagDomain.GetById(ctx, userId, tagId64)
	if err != nil {
		return nil, err
	}
	// 3. 实体转换响应体
	res := TagEntity2GetRes(tagEntity)
	// 4. 返回结果
	return res, nil
}

/*
 * Create tag
 * 创建标签
 */
func (tagApp *TagAppImpl) CreateTag(
	ctx context.Context,
	req *types.CreateTagReq,
) (*types.CreateTagRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换请求体
	createEntity := CreateReq2Entity(req)
	err := createEntity.IsValid()
	if err != nil {
		return nil, err
	}
	// 3. 调用域函数 - 创建标签
	tagEntity, err := tagApp.tagDomain.Create(ctx, userId, createEntity)
	if err != nil {
		return nil, err
	}
	// 4. 实体转换响应体并返回结果
	return TagEntity2CreateRes(tagEntity), nil
}

/*
 * Update tag
 * 更新标签信息
 */
func (tagApp *TagAppImpl) UpdateTag(
	ctx context.Context,
	req *types.UpdateTagReq,
) (*types.UpdateTagRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换 tagId
	tagId64, err := strconv.ParseInt(req.TagId, 10, 64)
	if err != nil {
		return nil, errors.New("标签 ID 格式错误")
	}
	// 3. 转换请求体
	tagEntity := UpdateReq2Entity(req)
	// 4. 调用域函数 - 更新标签信息
	err = tagApp.tagDomain.Update(ctx, userId, tagId64, tagEntity)
	if err != nil {
		return nil, err
	}
	// 5. 转换结果并返回
	return &types.UpdateTagRes{TagId: req.TagId}, nil
}

/*
 * Delete tag
 * 删除标签
 */
func (tagApp *TagAppImpl) DeleteTag(
	ctx context.Context,
	tagId string,
) (*types.DeleteTagRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 转换 tagId
	tagId64, err := strconv.ParseInt(tagId, 10, 64)
	if err != nil {
		return nil, errors.New("标签 ID 格式错误")
	}
	// 3. 调用域函数 - 删除标签信息
	err = tagApp.tagDomain.Delete(ctx, userId, tagId64)
	if err != nil {
		return nil, err
	}
	// 5. 转换结果并返回
	return &types.DeleteTagRes{TagId: tagId}, nil
}

/*
 * List tag
 * 获取所有标签信息
 */
func (tagApp *TagAppImpl) ListTag(
	ctx context.Context,
) (types.ListTagRes, error) {
	// 1. 获取用户 ID
	userId := iCtx.GetUserId(ctx)
	if userId <= 0 {
		return nil, errors.New("用户 ID 无效")
	}
	// 2. 调用域函数 - 删除标签信息
	tagEntities, err := tagApp.tagDomain.List(ctx, userId)
	if err != nil {
		return nil, err
	}
	// 5. 转换结果并返回
	return TagEntities2ListRes(tagEntities), nil
}

/*
 * Update tag preference
 * 更新标签偏好设置
 */
func (tagApp *TagAppImpl) UpdateTagPreference(
	ctx context.Context,
	tagId string,
	req *types.UpdateTagPreferenceReq,
) (*types.UpdateTagPreferenceRes, error) {
	panic("unimplemented")
}
