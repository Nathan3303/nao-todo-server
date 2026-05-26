package cron

import (
	"fmt"
	"naotodoserver/infrastructure/persistence/dbs"
	"naotodoserver/infrastructure/persistence/models"
	"time"
)

type DeleteDeactivedProjectJob struct {
	DayOffset int8
}

func NewDeleteDeactivedProjectJob(dayOffset int8) *DeleteDeactivedProjectJob {
	return &DeleteDeactivedProjectJob{DayOffset: dayOffset}
}

func (ddp *DeleteDeactivedProjectJob) Run() {
	tx := dbs.DB.Model(&models.Project{}).
		Where("deactived_at < ?", time.Now().AddDate(0, 0, -1*int(ddp.DayOffset))).
		Delete(&models.Project{})
	if tx.Error != nil {
		fmt.Println("删除软删除项目记录失败：" + tx.Error.Error())
		return
	}
	fmt.Printf("已删除软删除项目记录 %d 条\n", tx.RowsAffected)
}
