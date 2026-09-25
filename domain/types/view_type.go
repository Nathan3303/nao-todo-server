package types

// ViewType 视图类型
type ViewType string

const (
	ViewTypeTable  ViewType = "table"
	ViewTypeList   ViewType = "list"
	ViewTypeKanban ViewType = "kanban"
)
