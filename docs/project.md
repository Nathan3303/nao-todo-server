# 清单领域

实现清单的创建、查询、更新、删除，以及清单的偏好设置等操作。

## 应用层 - ProjectApp

1. 创建清单 CreateProject
    - 获取用户 ID
    - 定义清单实体
    - 调用域函数 - 创建清单
    - 清单实体转换响应体
    - 返回响应体

2. 查询清单 GetProject
    - 获取用户 ID
    - 获取清单 ID
    - 调用域函数 - 通过用户 ID 和清单 ID 获取清单记录
    - 清单实体转换响应体
    - 返回响应体

3. 更新清单 UpdateProject
    - 获取用户 ID
    - 获取清单 ID
    - 定义清单更新实体
    - 调用域函数 - 更新清单
    - 返回清单 ID

4. 删除清单 DeleteProject
    - 获取用户 ID
    - 获取清单 ID
    - 调用域函数 - 删除清单
    - 返回清单 ID

5. 恢复清单 RestoreProject
    - 获取用户 ID
    - 获取清单 ID
    - 调用域函数 - 恢复清单
    - 返回清单 ID

6. 归档清单 ArchiveProject
    - 获取用户 ID
    - 获取清单 ID
    - 调用域函数 - 归档清单
    - 返回清单 ID

7. 取消归档清单 UnarchiveProject
    - 获取用户 ID
    - 获取清单 ID
    - 调用域函数 - 取消归档清单
    - 返回清单 ID

8. 更新清单偏好 UpdateProjectPreference
    - 获取用户 ID
    - 获取清单 ID
    - 定义清单偏好更新实体
    - 调用域函数 - 更新清单偏好
    - 返回清单 ID

9. 获取清单列表 GetProjectsByUserId
    - 获取用户 ID
    - 调用域函数 - 通过用户 ID 获取清单列表
    - 清单实体转换响应体（遍历）
    - 返回响应体

## 领域层 - ProjectDomain

### 限界上下文

1. 创建清单
2. 通过用户 ID 和清单 ID 获取清单记录
3. 更新清单
4. 删除清单
5. 恢复清单
6. 归档清单
7. 取消归档清单
8. 更新清单偏好
9. 获取清单列表

### 实体 / 值对象

1. 清单实体 projectEntity
    - 清单 ID
    - 清单创建时间
    - 清单更新时间
    - 清单删除时间
    - 用户 ID
    - 清单名称
    - 清单描述
    - 清单归档日期
    - 清单偏好 | projectPreferenceVO

2. 清单偏好值对象 projectPreferenceVO
    - 清单偏好 ID
    - 清单偏好创建时间
    - 清单偏好更新时间
    - 清单偏好删除时间
    - 用户 ID
    - 清单 ID
    - 视图类型 ViewType
    - 任务获取选项 GetOptions
    - 列选项 Columns

### 仓库

1. 清单仓库 projectRepository
    - 创建清单
    - 通过用户 ID 和清单 ID 获取清单记录
    - 更新清单
    - 删除清单
    - 恢复清单
    - 归档清单
    - 取消归档清单
    - 更新清单偏好
    - 获取清单列表

## 基础层 - Infrastructure

1. 清单仓库实现
