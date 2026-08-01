package tag

import (
	"naotodoserver/application/idutil"
	"naotodoserver/application/tag/dto"
	"naotodoserver/domain/tag/entities"
	"naotodoserver/domain/tag/valueobjects"
	domaintypes "naotodoserver/domain/types"
	"time"
)

// TagEntityToGetRes 标签实体转换响应体
// @param tagEntity 标签实体
// @return 标签响应体
func TagEntityToGetRes(tagEntity *entities.Tag) *dto.GetTagRes {
	res := &dto.GetTagRes{}
	res.Id = idutil.FormatID(tagEntity.Id)
	res.CreatedAt = tagEntity.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = tagEntity.UpdatedAt.Format(time.RFC3339)
	res.DeletedAt = tagEntity.DeletedAt.ToString(time.RFC3339)
	res.Name = tagEntity.Name
	res.Description = tagEntity.Description
	res.Color = tagEntity.Color
	res.SortId = tagEntity.SortId
	return res
}

// CreateTagReqToValueObject 创建标签请求体转换值对象
// @param createTagReq 创建标签请求体
// @return 创建标签值对象
func CreateTagReqToValueObject(createTagReq *dto.CreateTagReq) (*valueobjects.CreateTag, error) {
	return valueobjects.NewCreateTag(
		createTagReq.Name,
		createTagReq.Description,
		createTagReq.Color,
	)
}

// TagEntityToCreateRes 标签实体转换创建响应体
// @param tagEntity 标签实体
// @return 创建标签响应体
func TagEntityToCreateRes(tagEntity *entities.Tag) *dto.CreateTagRes {
	res := &dto.CreateTagRes{}
	res.Id = idutil.FormatID(tagEntity.Id)
	res.CreatedAt = tagEntity.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = tagEntity.UpdatedAt.Format(time.RFC3339)
	res.DeletedAt = tagEntity.DeletedAt.ToString(time.RFC3339)
	res.Name = tagEntity.Name
	res.Description = tagEntity.Description
	res.Color = tagEntity.Color
	res.SortId = tagEntity.SortId
	return res
}

// UpdateTagReqToValueObject 更新标签请求体转换值对象
// @param updateTagReq 更新标签请求体
// @return 更新标签值对象
func UpdateTagReqToValueObject(updateTagReq *dto.UpdateTagReq) (*valueobjects.UpdateTag, error) {
	return valueobjects.NewUpdateTag(
		updateTagReq.Name,
		updateTagReq.Description,
		updateTagReq.Color,
		updateTagReq.SortId,
	)
}

// TagEntitiesToGetResList 标签实体列表转换响应体列表
// @param tagEntities 标签实体列表
// @return 标签响应体列表
func TagEntitiesToGetResList(tagEntities []*entities.Tag) []*dto.GetTagRes {
	getResList := []*dto.GetTagRes{}
	for _, entity := range tagEntities {
		getResList = append(getResList, TagEntityToGetRes(entity))
	}
	return getResList
}

// TagPreferenceEntityToGetRes 标签偏好设置实体转换响应体
// @param tagPreferenceEntity 标签偏好设置实体
// @return 标签偏好设置响应体
func TagPreferenceEntityToGetRes(
	tagPreferenceEntity *entities.TagPreference,
) *dto.GetTagPreferenceRes {
	var res dto.GetTagPreferenceRes
	res.Id = idutil.FormatID(tagPreferenceEntity.Id)
	res.CreatedAt = tagPreferenceEntity.CreatedAt.Format(time.RFC3339)
	res.UpdatedAt = tagPreferenceEntity.UpdatedAt.Format(time.RFC3339)
	res.DeletedAt = tagPreferenceEntity.DeletedAt.ToString(time.RFC3339)
	res.TagId = idutil.FormatID(tagPreferenceEntity.TagId)
	res.ViewType = tagPreferenceEntity.ViewType
	res.GetOptions = tagPreferenceEntity.GetOptions
	res.Columns = tagPreferenceEntity.Columns
	return &res
}

// UpdateTagPreferenceReqToValueObject 更新标签偏好设置请求体转换值对象
// @param updateTagPreferenceReq 更新标签偏好设置请求体
// @return 更新标签偏好设置值对象
func UpdateTagPreferenceReqToValueObject(
	updateTagPreferenceReq *dto.UpdateTagPreferenceReq,
) (*valueobjects.SaveTagPreference, error) {
	return valueobjects.NewSaveTagPreference(
		domaintypes.ViewType(updateTagPreferenceReq.ViewType),
		updateTagPreferenceReq.GetOptions,
		updateTagPreferenceReq.Columns,
	)
}

// BatchUpdateTagReqToValueObjects 批量更新标签请求体转换值对象列表
// @param req 批量更新标签请求体
// @return 批量更新标签值对象列表
// @return error 错误信息
func BatchUpdateTagReqToValueObjects(
	req *dto.BatchUpdateTagReq,
) ([]*valueobjects.BatchUpdateTag, error) {
	batchVOs := make([]*valueobjects.BatchUpdateTag, 0, len(req.Tags))
	for _, tag := range req.Tags {
		id, err := idutil.ParseID(tag.Id)
		if err != nil {
			return nil, err
		}
		vo, err := valueobjects.NewBatchUpdateTag(
			id,
			tag.Name,
			tag.Description,
			tag.Color,
			tag.SortId,
		)
		if err != nil {
			return nil, err
		}
		batchVOs = append(batchVOs, vo)
	}
	return batchVOs, nil
}
