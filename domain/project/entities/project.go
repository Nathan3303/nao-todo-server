package entities

import (
	"naotodoserver/domain/project/vo"
	"time"
)

type Project struct {
	Id          int64                 `json:"id"`
	UserId      int64                 `json:"userId"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	ArchivedAt  *time.Time            `json:"archivedAt"`
	Preference  *vo.ProjectPreference `json:"preference"`
}

func (p *Project) IsIdValid() bool {
	return p != nil && p.Id > 0
}

func (p *Project) IsUserIdValid() bool {
	return p != nil && p.UserId > 0
}

func (p *Project) IsValid() bool {
	return p.IsIdValid() && p.IsUserIdValid()
}
