
# Project Model 添加 deactivatedAt 属性实现计划

## 概述
为 Project Model 添加 `deactivatedAt` 属性作为软删除标记。对内它是软删除标记，对外它表示项目已删除状态，同时保留 `deletedAt` 用于永久删除但保留数据的情况。

**功能设计**:
- 只对外暴露软删除 API
- 被软删除的数据会在 30 天后通过后台任务硬删除
- 定时任务参考 `deleteInactiveUser.go` 的实现方式

## 现状分析

### 当前实现
- 使用 GORM 内置的 `DeletedAt` 字段（gorm.DeletedAt）实现软删除
- `Project` 实体已有 `DeletedAt` 字段
- Repository 中的 `Delete` 方法使用 GORM 的 `Delete()` 自动设置 deleted_at
- Repository 中的 `Restore` 方法使用 `Unscoped()` 恢复记录

### 需要修改的文件清单
1. `domain/project/entities/project.go` - 实体定义
2. `infrastructure/persistence/models/project.go` - Project 模型
3. `infrastructure/persistence/project/converters.go` - 转换器
4. `infrastructure/persistence/project/repoImpl.go` - Repository 实现
5. `application/project/converters.go` - 应用层转换器
6. `interfaces/types/project.go` - 接口类型定义
7. `infrastructure/cron/deleteInactiveProject.go` - 新建定时任务文件

## 实现步骤

### 步骤 1: 修改实体定义
**文件**: `domain/project/entities/project.go`

- 保留 `DeletedAt *time.Time` 字段（用于永久删除）
- 添加 `DeactivatedAt *time.Time` 字段（用于软删除）

### 步骤 2: 修改 Project 模型
**文件**: `infrastructure/persistence/models/project.go`

- 保留 ModelBase 中的 DeletedAt（保持不变）
- 添加 `DeactivatedAt *time.Time` 字段
- 配置 gorm tag: `gorm:"index"`

### 步骤 3: 修改转换器
**文件**: `infrastructure/persistence/project/converters.go`

- 更新 `Model2Entity` 函数，添加 DeactivatedAt 字段转换
- 更新 `Entity2Model` 函数，添加 DeactivatedAt 字段转换

### 步骤 4: 修改 Repository 实现
**文件**: `infrastructure/persistence/project/repoImpl.go`

- 修改 `Delete` 方法：不再使用 GORM Delete，而是更新 deactivated_at 为当前时间
- 修改 `Restore` 方法：将 deactivated_at 设置为 nil
- 修改 `GetById` 方法：添加条件 `deactivated_at IS NULL`
- 修改 `GetByUserId` 方法：添加条件 `deactivated_at IS NULL`

### 步骤 5: 修改应用层转换器
**文件**: `application/project/converters.go`

- 更新 `ProjectEntityToCreateRes` 函数，添加 DeactivatedAt 字段
- 更新 `ProjectEntityToGetRes` 函数，添加 DeactivatedAt 字段

### 步骤 6: 修改接口类型定义
**文件**: `interfaces/types/project.go`

- 更新 `CreateProjectRes`，添加 `DeactivatedAt *time.Time` 字段
- 更新 `GetProjectRes`，添加 `DeactivatedAt *time.Time` 字段

### 步骤 7: 新建定时任务文件
**文件**: `infrastructure/cron/deleteInactiveProject.go`

- 参考 `deleteInactiveUser.go` 的实现
- 创建 `DeleteDeactivedProjectJob` 结构体
- 实现 `Run()` 方法，删除 30 天前软删除的项目

## 详细变更内容

### domain/project/entities/project.go
```go
type Project struct {
    Id           int64
    UserId       int64
    Name         string
    Description  string
    ArchivedAt   *time.Time
    CreatedAt    time.Time
    UpdatedAt    time.Time
    DeletedAt    *time.Time    // 保留，用于永久删除
    DeactivatedAt *time.Time  // 新增，用于软删除
}
```

### infrastructure/persistence/models/project.go
```go
type Project struct {
    ModelBase    // 保留 DeletedAt
    UserId       int64          `gorm:"not null"`
    Name         string         `gorm:"size:128"`
    Description  string         `gorm:"size:256"`
    ArchivedAt   *time.Time     `gorm:"null"`
    DeactivatedAt *time.Time    `gorm:"index"`  // 新增字段
    Preference   *ProjectPreference `gorm:"foreignKey:ProjectId;constraint:OnDelete:CASCADE;"`
}
```

### infrastructure/persistence/project/repoImpl.go 关键修改
```go
// Delete 改为设置 deactivated_at（软删除）
func (projectRepo *ProjectRepoImpl) Delete(...) error {
    tx := projectRepo.db.WithContext(ctx).
        Model(&models.Project{}).
        Where(&whereCond).
        Update("deactivated_at", time.Now())
    // ...
}

// Restore 恢复 deactivated_at 为 nil
func (projectRepo *ProjectRepoImpl) Restore(...) error {
    tx := projectRepo.db.WithContext(ctx).
        Model(&models.Project{}).
        Where(&whereCond).
        Update("deactivated_at", nil)
    // ...
}

// GetById 和 GetByUserId 添加过滤条件
Where("user_id = ? AND id = ? AND deactivated_at IS NULL", ...)
```

### infrastructure/cron/deleteInactiveProject.go
```go
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
		Where("deactivated_at &lt; ?", time.Now().AddDate(0, 0, -1*int(ddp.DayOffset))).
		Delete(&models.Project{})
	if tx.Error != nil {
		fmt.Println("删除软删除项目记录失败：" + tx.Error.Error())
		return
	}
	fmt.Printf("已删除软删除项目记录 %d 条\n", tx.RowsAffected)
}
```

## 字段语义说明

### 对内（实现层面）
- `deactivated_at`: 软删除标记，数据保留在数据库
- `deleted_at`: 永久删除标记（GORM 软删除）

### 对外（API 层面）
- `deactivated_at`: 表示项目已删除，用户不可见
- 普通查询只返回 `deactivated_at IS NULL` 的项目
- 不提供硬删除 API

## 定时任务
- 新建 `deleteInactiveProject.go` 文件
- 参考 `deleteInactiveUser.go` 实现
- 定期扫描 `deactivated_at` 超过 30 天的项目
- 自动执行硬删除（设置 `deleted_at`）

## 注意事项
- 查询时自动过滤 deactivated_at 不为空的项目
- 需要确保数据库迁移时正确添加新字段
- ProjectPreference 保持不变
- 硬删除只通过定时任务执行，不对外暴露 API
- 注意字段名是 `deactivated_at`（不是 `deactived_at`）
