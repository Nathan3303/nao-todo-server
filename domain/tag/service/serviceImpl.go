package service

import (
	"context"
	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/repositories"
	"naotodoserver/domain/tag/vo"
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

/*
 * Get tag by id
 * 根据标签ID获取标签信息
 */
func (tagDomain *TagDomainImpl) GetById(
	ctx context.Context,
	userId int64,
	tagId int64,
) (*entities.Tag, error) {
	return tagDomain.tagRepo.GetById(ctx, userId, tagId)
}

/*
 * Create tag
 * 创建标签
 */
func (tagDomain *TagDomainImpl) Create(
	ctx context.Context,
	userId int64,
	createEntity *entities.Tag,
) (*entities.Tag, error) {
	createEntity.UserId = userId
	// createEntity.Preference = vo.MakeDefaultTagPreference()
	return tagDomain.tagRepo.Create(ctx, createEntity)
}

/*
 * Update tag
 * 更新标签信息
 */
func (tagDomain *TagDomainImpl) Update(
	ctx context.Context,
	userId int64,
	tagId int64,
	updateEntity *entities.Tag,
) error {
	whereEntity := &entities.Tag{UserId: userId, Id: tagId}
	return tagDomain.tagRepo.Update(ctx, whereEntity, updateEntity)
}

/*
 * Delete tag
 * 删除标签
 */
func (tagDomain *TagDomainImpl) Delete(
	ctx context.Context,
	userId int64,
	tagId int64,
) error {
	whereEntity := &entities.Tag{UserId: userId, Id: tagId}
	return tagDomain.tagRepo.Delete(ctx, whereEntity)
}

/*
 * List tag
 * 获取所有标签
 */
func (tagDomain *TagDomainImpl) List(
	ctx context.Context,
	userId int64,
) ([]*entities.Tag, error) {
	whereEntity := &entities.Tag{UserId: userId}
	return tagDomain.tagRepo.Get(ctx, whereEntity)
}

/*
 * Get tag preference
 * 获取标签偏好
 */
func (tagDomain *TagDomainImpl) GetPreference(
	ctx context.Context,
	userId int64,
	tagId int64,
) (*vo.TagPreference, error) {
	return tagDomain.preferenceRepo.Get(ctx, &vo.TagPreference{
		UserId: userId,
		TagId:  tagId,
	})
}

/*
 * Update tag preference
 * 更新标签偏好
 */
func (tagDomain *TagDomainImpl) UpdatePreference(
	ctx context.Context,
	userId int64,
	tagId int64,
	preference *vo.TagPreference,
) error {
	preference.UserId = userId
	preference.TagId = tagId
	_, err := tagDomain.preferenceRepo.Save(ctx, preference)
	if err != nil {
		return err
	}
	return nil
}
