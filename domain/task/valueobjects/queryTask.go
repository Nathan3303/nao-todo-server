package valueobjects

type QueryTask struct {
	UserId       int64
	ProjectId    int64
	TagId        string
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
}

func (queryTask *QueryTask) Validate() error {
	return nil
}

func NewQueryTask(
	userId int64,
	projectid int64,
	tagId string,
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
		ProjectId:    projectid,
		TagId:        tagId,
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
