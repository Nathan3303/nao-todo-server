package cron

import (
	"fmt"
	"naotodoserver/infrastructure/persistence/dbs"
	"naotodoserver/infrastructure/persistence/models"
	"time"
)

type DeleteDeactivedUserJob struct {
	DayOffset int8
}

func NewDeleteDeactivedUserJob(dayOffset int8) *DeleteDeactivedUserJob {
	return &DeleteDeactivedUserJob{DayOffset: dayOffset}
}

func (ddu *DeleteDeactivedUserJob) Run() {
	tx := dbs.DB.Model(&models.User{}).
		Where("deactived_at < ?", time.Now().AddDate(0, 0, -1*int(ddu.DayOffset))).
		Delete(&models.User{})
	if tx.Error != nil {
		fmt.Println("删除注销用户记录失败：" + tx.Error.Error())
		return
	}
	fmt.Printf("已删除注销用户记录 %d 条\n", tx.RowsAffected)
}
