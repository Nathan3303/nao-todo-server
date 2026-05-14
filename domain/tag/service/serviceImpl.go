package service

import (
	"context"
	"fmt"
	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/repositories"
	"naotodoserver/domain/tag/valueobjects"
)

// 标签域注册函数
func NewTagDomain(
	tagRepo repositories.TagRepository,
	preferenceRepo repositories.TagPreference,
) TagDomain {
	return &TagDomainImpl{
		tagRepo:        tagRepo,
		preferenceRepo: preferenceRepo,
	}
}

// 根据标签ID获取标签信息
// @param ctx 上下文
// @param userId 用户ID
// @param tagId 标签ID
// @return *entities.Tag 标签信息
// @return error 校验失败返回错误，否则返回 nil
func (tagDomain *TagDomainImpl) GetById(
	ctx context.Context,
	userId int64,
	tagId int64,
) (*entities.Tag, error) {
	return tagDomain.tagRepo.GetById(ctx, userId, tagId)
}

// 创建标签
// @param ctx 上下文
// @param userId 用户ID
// @param createTagValueObject 创建标签值对象
// @return *entities.Tag 创建的标签信息
// @return error 校验失败返回错误，否则返回 nil
func (tagDomain *TagDomainImpl) Create(
	ctx context.Context,
	userId int64,
	createTagValueObject *valueobjects.CreateTag,
) (*entities.Tag, error) {
	// 设置排序 ID
	createTagValueObject.SortId = tagDomain.tagRepo.GetMaxSortId(ctx, userId) + 1
	// 创建标签
	tagEntity, err := tagDomain.tagRepo.Create(ctx, userId, createTagValueObject)
	if err != nil {
		return nil, err
	}
	// 创建标签基础偏好
	tagPreferenceValueObject, err := valueobjects.NewSaveTagPreference(
		"table",
		fmt.Sprintf("{\"tagId\": \"%d\"}", tagEntity.Id),
		"{}",
	)
	if err != nil {
		return nil, err
	}
	// 保存标签基础偏好
	err = tagDomain.preferenceRepo.Save(
		ctx,
		userId,
		tagEntity.Id,
		tagPreferenceValueObject,
	)
	if err != nil {
		return nil, err
	}
	// 返回标签信息
	return tagEntity, nil
}

// 更新标签
// @param ctx 上下文
// @param userId 用户ID
// @param tagId 标签ID
// @param updateTagValueObject 更新标签值对象
// @return error 校验失败返回错误，否则返回 nil
func (tagDomain *TagDomainImpl) Update(
	ctx context.Context,
	userId int64,
	tagId int64,
	updateTagValueObject *valueobjects.UpdateTag,
) error {
	return tagDomain.tagRepo.Update(ctx, userId, tagId, updateTagValueObject)
}

// 删除标签
// @param ctx 上下文
// @param userId 用户ID
// @param tagId 标签ID
// @return error 校验失败返回错误，否则返回 nil
func (tagDomain *TagDomainImpl) Delete(
	ctx context.Context,
	userId int64,
	tagId int64,
) error {
	// 删除标签
	err := tagDomain.tagRepo.Delete(ctx, userId, tagId)
	if err != nil {
		return err
	}
	// 删除标签偏好
	return tagDomain.preferenceRepo.Delete(ctx, userId, tagId)
}

// 获取标签列表
// @param ctx 上下文
// @param userId 用户ID
// @return []*entities.Tag 标签列表
// @return error 校验失败返回错误，否则返回 nil
func (tagDomain *TagDomainImpl) List(
	ctx context.Context,
	userId int64,
) ([]*entities.Tag, error) {
	return tagDomain.tagRepo.Get(ctx, userId)
}

// 批量更新标签
func (tagDomain *TagDomainImpl) BatchUpdate(
	ctx context.Context,
	userId int64,
	batchUpdateTags []*valueobjects.BatchUpdateTag,
) ([]*entities.Tag, error) {
	return tagDomain.tagRepo.BatchUpdate(ctx, userId, batchUpdateTags)
}

// 获取标签偏好
// @param ctx 上下文
// @param userId 用户ID
// @param tagId 标签ID
// @return *entities.TagPreference 标签偏好
// @return error 校验失败返回错误，否则返回 nil
func (tagDomain *TagDomainImpl) GetPreference(
	ctx context.Context,
	userId int64,
	tagId int64,
) (*entities.TagPreference, error) {
	return tagDomain.preferenceRepo.Get(ctx, userId, tagId)
}

// 更新标签偏好
// @param ctx 上下文
// @param userId 用户ID
// @param tagId 标签ID
// @param saveTagPreferenceValueObject 更新标签偏好值对象
// @return error 校验失败返回错误，否则返回 nil
func (tagDomain *TagDomainImpl) UpdatePreference(
	ctx context.Context,
	userId int64,
	tagId int64,
	saveTagPreferenceValueObject *valueobjects.SaveTagPreference,
) error {
	return tagDomain.preferenceRepo.Save(ctx, userId, tagId, saveTagPreferenceValueObject)
}
