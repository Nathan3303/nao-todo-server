# Event 域批量更新 API 实现计划

## 1. 项目架构分析

本项目采用分层架构设计，主要分为以下层次：

* **接口层 (Interfaces Layer)**: 包含控制器、路由和类型定义

* **应用层 (Application Layer)**: 包含应用服务接口和实现

* **领域层 (Domain Layer)**: 包含实体、值对象、服务和存储库接口

* **基础设施层 (Infrastructure Layer)**: 包含存储库实现和其他基础设施代码

## 2. 实现计划步骤

### 步骤 1: 添加接口层类型定义

在 `interfaces/types/event.go` 文件中添加批量更新事件的请求和响应类型。

### 步骤 2: 更新应用层接口和实现

在 `application/event/app.go` 中添加应用服务接口，并在 `appImpl.go` 中实现该接口。

### 步骤 3: 更新领域层

* 在 `domain/event/repositories/event.go` 中添加存储库接口

* 在 `domain/event/service/service.go` 和 `serviceImpl.go` 中添加领域服务方法

* 在 `domain/event/valueobjects/` 目录下创建批量更新的值对象

### 步骤 4: 实现基础设施层

在 `infrastructure/persistence/event/repoImpl.go` 中实现存储库接口的批量更新方法。

### 步骤 5: 添加接口层控制器

在 `interfaces/controllers/event.go` 中添加批量更新事件的控制器方法。

### 步骤 6: 添加路由

在 `interfaces/routers/eventRouter.go` 中添加对应的路由配置。

## 3. 文件修改清单

| 文件名                                             | 修改内容           |
| ----------------------------------------------- | -------------- |
| `interfaces/types/event.go`                     | 添加批量更新的请求和响应类型 |
| `application/event/app.go`                      | 添加应用服务接口       |
| `application/event/appImpl.go`                  | 实现应用服务接口       |
| `domain/event/repositories/event.go`            | 添加存储库接口        |
| `domain/event/service/service.go`               | 添加领域服务接口       |
| `domain/event/service/serviceImpl.go`           | 实现领域服务接口       |
| `domain/event/valueobjects/batchUpdateEvent.go` | 创建批量更新值对象      |
| `infrastructure/persistence/event/repoImpl.go`  | 实现存储库接口        |
| `interfaces/controllers/event.go`               | 添加控制器方法        |
| `interfaces/routers/eventRouter.go`             | 添加路由配置         |

## 4. 接口设计

### 请求路径

`PUT /api/events`

### 请求体格式

```json
{
  "events": [
    {
      "id": "1",
      "name": "完成项目报告",
      "description": "撰写项目开发报告",
      "isDone": true,
      "sortId": 1
    },
    {
      "id": "2",
      "name": "准备会议材料",
      "description": "准备团队周会的演示材料",
      "isDone": false,
      "sortId": 2
    }
  ]
}
```

### 响应体格式

```json
{
  "code": 50060,
  "message": "批量更新检查事项成功",
  "data": {
    "updatedCount": 2,
    "events": [
      {
        "id": "1",
        "taskId": "101",
        "name": "完成项目报告",
        "description": "撰写项目开发报告",
        "isDone": true,
        "sortId": 1,
        "createdAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-02T00:00:00Z"
      },
      {
        "id": "2",
        "taskId": "101",
        "name": "准备会议材料",
        "description": "准备团队周会的演示材料",
        "isDone": false,
        "sortId": 2,
        "createdAt": "2024-01-01T00:00:00Z",
        "updatedAt": "2024-01-02T00:00:00Z"
      }
    ]
  }
}
```

## 5. 错误处理

API 将处理以下错误场景：

* 无效的请求参数

* 用户身份验证失败

* 事件 ID 格式错误

* 数据库操作失败

* 权限验证失败

## 6. 设计原则

* 遵循项目现有架构风格

* 使用值对象进行数据验证

* 实现领域服务层的业务逻辑

* 在基础设施层实现数据持久化

* 提供清晰的接口文档

## 7. 测试计划

* 单元测试：测试每个层次的功能

* 集成测试：测试整个 API 调用链

* 性能测试：测试批量更新的性能

## 8. 依赖关系

* 项目使用 Gin 框架

* 使用 GORM 进行数据库操作

* 使用 JWT 进行身份验证

* 依赖于现有项目结构

