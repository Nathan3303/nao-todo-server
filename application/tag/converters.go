package tag

import (
	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/valueobjects"
	"naotodoserver/interfaces/types"
	"strconv"
)

// 标签实体转换响应体
// @param tagEntity 标签实体
// @return 标签响应体
func TagEntityToGetRes(tagEntity *entities.Tag) *types.GetTagRes {
	res := &types.GetTagRes{}
	res.Id = strconv.FormatInt(tagEntity.Id, 10)
	res.Name = tagEntity.Name
	res.Description = tagEntity.Description
	res.Color = tagEntity.Color
	res.SortId = tagEntity.SortId
	res.CreatedAt = tagEntity.CreatedAt
	res.UpdatedAt = tagEntity.UpdatedAt
	return res
}

// 创建标签请求体转换值对象
// @param createTagReq 创建标签请求体
// @return 创建标签值对象
func CreateTagReqToValueObject(createTagReq *types.CreateTagReq) (*valueobjects.CreateTag, error) {
	return valueobjects.NewCreateTag(
		createTagReq.Name,
		createTagReq.Description,
		createTagReq.Color,
	)
}

// 标签实体转换创建响应体
// @param tagEntity 标签实体
// @return 创建标签响应体
func TagEntityToCreateRes(tagEntity *entities.Tag) *types.CreateTagRes {
	res := &types.CreateTagRes{}
	res.Id = strconv.FormatInt(tagEntity.Id, 10)
	res.Name = tagEntity.Name
	res.Description = tagEntity.Description
	res.Color = tagEntity.Color
	res.SortId = tagEntity.SortId
	res.CreatedAt = tagEntity.CreatedAt
	res.UpdatedAt = tagEntity.UpdatedAt
	return res
}

// 更新标签请求体转换值对象
// @param updateTagReq 更新标签请求体
// @return 更新标签值对象
func UpdateTagReqToValueObject(updateTagReq *types.UpdateTagReq) (*valueobjects.UpdateTag, error) {
	return valueobjects.NewUpdateTag(
		updateTagReq.Name,
		updateTagReq.Description,
		updateTagReq.Color,
		updateTagReq.SortId,
	)
}

// 标签实体列表转换响应体列表
// @param tagEntities 标签实体列表
// @return 标签响应体列表
func TagEntitiesToGetResList(tagEntities []*entities.Tag) []*types.GetTagRes {
	getResList := []*types.GetTagRes{}
	for _, entity := range tagEntities {
		getResList = append(getResList, TagEntityToGetRes(entity))
	}
	return getResList
}

// 标签偏好设置实体转换响应体
// @param tagPreferenceEntity 标签偏好设置实体
// @return 标签偏好设置响应体
func TagPreferenceEntityToGetRes(
	tagPreferenceEntity *entities.TagPreference,
) *types.GetTagPreferenceRes {
	return &types.GetTagPreferenceRes{
		Id:         strconv.FormatInt(tagPreferenceEntity.Id, 10),
		TagId:      strconv.FormatInt(tagPreferenceEntity.TagId, 10),
		ViewType:   tagPreferenceEntity.ViewType,
		GetOptions: tagPreferenceEntity.GetOptions,
		Columns:    tagPreferenceEntity.Columns,
		CreatedAt:  tagPreferenceEntity.CreatedAt,
		UpdatedAt:  tagPreferenceEntity.UpdatedAt,
	}
}

// 更新标签偏好设置请求体转换值对象
// @param updateTagPreferenceReq 更新标签偏好设置请求体
// @return 更新标签偏好设置值对象
func UpdateTagPreferenceReqToValueObject(
	updateTagPreferenceReq *types.UpdateTagPreferenceReq,
) (*valueobjects.SaveTagPreference, error) {
	return valueobjects.NewSaveTagPreference(
		updateTagPreferenceReq.ViewType,
		updateTagPreferenceReq.GetOptions,
		updateTagPreferenceReq.Columns,
	)
}

// 批量更新标签请求体转换值对象
func BatchUpdateTagReqToValueObjects(
	req *types.BatchUpdateTagReq,
) ([]*valueobjects.BatchUpdateTag, error) {
	batchVOs := make([]*valueobjects.BatchUpdateTag, 0, len(req.Tags))
	for _, tag := range req.Tags {
		id, err := strconv.ParseInt(tag.Id, 10, 64)
		if err != nil {
			return nil, err
		}
		vo, err := valueobjects.NewBatchUpdateTag(id, tag.Name, tag.Description, tag.Color, tag.SortId)
		if err != nil {
			return nil, err
		}
		batchVOs = append(batchVOs, vo)
	}
	return batchVOs, nil
}
