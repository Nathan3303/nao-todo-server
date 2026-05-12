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
	cutoff := time.Now().AddDate(0, 0, -1*int(ddu.DayOffset))
	// 查询需要删除的用户ID
	var userIds []int64
	tx := dbs.DB.Model(&models.User{}).
		Where("deactived_at < ?", cutoff).
		Pluck("id", &userIds)
	if tx.Error != nil {
		fmt.Println("查询注销用户记录失败：" + tx.Error.Error())
		return
	}
	if len(userIds) > 0 {
		// 先软删除关联的 UserConfig
		dbs.DB.Model(&models.UserConfig{}).
			Where("user_id IN ?", userIds).
			Delete(&models.UserConfig{})
	}
	// 软删除用户
	tx = dbs.DB.Model(&models.User{}).
		Where("id IN ?", userIds).
		Delete(&models.User{})
	if tx.Error != nil {
		fmt.Println("删除注销用户记录失败：" + tx.Error.Error())
		return
	}
	fmt.Printf("已删除注销用户记录 %d 条\n", tx.RowsAffected)
}
