package service

import (
	"context"
	"fmt"
	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/repositories"
	"naotodoserver/domain/tag/valueobjects"
)

// NewTagDomain 标签域注册函数
func NewTagDomain(
	tagRepo repositories.TagRepository,
	preferenceRepo repositories.TagPreference,
) TagDomain {
	return &TagDomainImpl{
		tagRepo:        tagRepo,
		preferenceRepo: preferenceRepo,
	}
}

// Create 创建标签
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
	// 幂等创建：客户端指定 id 时走 upsert（LWW + create 冲突检测）
	tagEntity, created, err := tagDomain.tagRepo.Upsert(ctx, userId, createTagValueObject)
	if err != nil {
		return nil, err
	}
	// 仅首次新建时初始化基础偏好；覆盖/重试场景保留用户已有偏好
	if !created {
		return tagEntity, nil
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

// Delete 删除标签
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
