package entities

import (
	"errors"
	"naotodoserver/domain/tag/vo"
)

type Tag struct {
	Id          int64             `json:"id"`
	UserId      int64             `json:"userId"`
	Name        string            `json:"name"`
	Description string            `json:"description"`
	Color       string            `json:"color"`
	Preference  *vo.TagPreference `json:"preference"`
}

func (tag *Tag) IsIdValid() bool {
	return tag != nil && tag.Id > 0
}

func (tag *Tag) IsNameValid() bool {
	return tag != nil && len(tag.Name) > 2 && len(tag.Name) <= 32
}

func (tag *Tag) IsDescriptionValid() bool {
	return tag != nil && len(tag.Description) <= 256
}

func (tag *Tag) IsColorValid() bool {
	return tag != nil && len(tag.Color) > 0 && len(tag.Color) <= 16
}

func (tag *Tag) IsValid() error {
	if tag == nil {
		return errors.New("标签实体为空")
	}
	if !tag.IsNameValid() {
		return errors.New("标签名称无效")
	}
	if !tag.IsDescriptionValid() {
		return errors.New("标签描述无效")
	}
	if !tag.IsColorValid() {
		return errors.New("标签颜色无效")
	}
	return nil
}
