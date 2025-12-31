package vo

type TaskQuery struct {
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
