package valueobjects

import "time"

type QueryTask struct {
	UserId       int64
	ParentTaskId int64
	ProjectIds   []int64
	TagIds       []string
	Name         string
	Description  string
	State        string
	Priority     string
	StartAt      string
	EndAt        string
	DeletedAt    string
	ArchivedAt   string
	StarMarkAt   string
	GivenUpAt    string
	IsDeleted    bool
	IsArchived   bool
	IsStarMarked bool
	IsGivenUp    bool
	Page         int
	Limit        int
	RelativeDate string
	Sort         string
	// UpdatedAt 增量同步游标：updated_at >= 该值；零值表示普通查询
	UpdatedAt time.Time
}

func (queryTask *QueryTask) Validate() error {
	return nil
}

func NewQueryTask(
	userId int64,
	parentTaskId int64,
	projectIds []int64,
	tagIds []string,
	name string,
	description string,
	state string,
	priority string,
	startAt string,
	endAt string,
	deletedAt string,
	archivedAt string,
	starMarkAt string,
	givenUpAt string,
	isDeleted bool,
	isArchived bool,
	isStarMarked bool,
	isGivenUp bool,
	page int,
	limit int,
	relativeDate string,
	sort string,
) (*QueryTask, error) {
	vo := &QueryTask{
		UserId:       userId,
		ParentTaskId: parentTaskId,
		ProjectIds:   projectIds,
		TagIds:       tagIds,
		Name:         name,
		Description:  description,
		State:        state,
		Priority:     priority,
		StartAt:      startAt,
		EndAt:        endAt,
		DeletedAt:    deletedAt,
		ArchivedAt:   archivedAt,
		StarMarkAt:   starMarkAt,
		GivenUpAt:    givenUpAt,
		IsDeleted:    isDeleted,
		IsArchived:   isArchived,
		IsStarMarked: isStarMarked,
		IsGivenUp:    isGivenUp,
		Page:         page,
		Limit:        limit,
		RelativeDate: relativeDate,
		Sort:         sort,
	}
	err := vo.Validate()
	if err != nil {
		return nil, err
	}
	return vo, nil
}
