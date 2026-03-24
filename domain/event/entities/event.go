package entities

import "errors"

type Event struct {
	Id          int64
	UserId      int64
	TaskId      int64
	Name        string
	Description string
	IsDone      bool
	SortId      uint32
}

func (e *Event) IsNameValid() bool {
	return e != nil &&
		len(e.Name) > 0 && len(e.Name) <= 128
}

func (e *Event) IsDescriptionValid() bool {
	return e != nil &&
		len(e.Description) <= 256
}

func (e *Event) IsValid() error {
	if !e.IsNameValid() {
		return errors.New("检查事项名称无效")
	}
	if !e.IsDescriptionValid() {
		return errors.New("检查事项描述无效")
	}
	return nil
}
