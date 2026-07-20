# 接口错误响应统一以 HTTP 200 返回

## 需求分析

用户要求：

1. 所有接口返回的错误都应该以 **HTTP 200** 状态码返回
2. 错误状态由返回结构体中的 `Code`、`Message` 以及 `Error`（非公开）字段标识
3. `Code` 以 `0` 结尾表示成功
4. `Message` 用于传递给用户错误/成功信息

## 当前状态分析

### 现有响应结构

[responseData.go](file:///C:/Users/LEE19/Projects/nao-todo-server/interfaces/types/responseData.go) 中已定义了响应结构体：

```go
type ResponseData struct {
    Code       int         `json:"code"`
    Message    string      `json:"message,omitempty"`
    Data       any         `json:"data,omitempty"`
    Pagination *Pagination `json:"pagination,omitempty"`
    Status     int         `json:"status,omitempty"`
    Error      string      `json:"error,omitempty"` // 错误用
}
```

响应结构已满足需求，包含 `Code`、`Message`、`Error` 字段。

### 现有响应函数

[responser.go](file:///C:/Users/LEE19/Projects/nao-todo-server/interfaces/controllers/responser.go) 中定义了响应函数：

```go
func Success(ctx *gin.Context, resData types.ResponseData) {
    EndWithStatus(ctx, http.StatusOK, resData)  // HTTP 200 ✅
}

func Failure(ctx *gin.Context, resData types.ResponseData) {
    EndWithStatus(ctx, http.StatusBadRequest, resData)  // HTTP 400 ❌ 需要改为 200
}
```

**问题**：`Failure` 函数使用 `http.StatusBadRequest`（HTTP 400），需要改为 `http.StatusOK`（HTTP 200）。

## 修改计划

### 修改文件

| 文件                                                                                                  | 修改内容                                                        |
| --------------------------------------------------------------------------------------------------- | ----------------------------------------------------------- |
| [responser.go](file:///C:/Users/LEE19/Projects/nao-todo-server/interfaces/controllers/responser.go) | 将 `Failure` 函数中的 `http.StatusBadRequest` 改为 `http.StatusOK` |

### 修改步骤

1. 修改 `responser.go` 中的 `Failure` 函数：

   * 将 `EndWithStatus(ctx, http.StatusBadRequest, resData)` 改为 `EndWithStatus(ctx, http.StatusOK, resData)`

2. 验证修改：

   * 检查所有控制器中使用 `Failure` 的地方是否仍然正确工作

   * 确保所有错误响应都返回 HTTP 200

## 预期效果

修改后，所有接口响应（包括成功和失败）都将返回 HTTP 200 状态码，错误信息通过响应体中的 `Code`、`Message` 和 `Error` 字段传递。

### 成功响应示例

```json
{
    "code": 20000,
    "message": "获取清单成功",
    "data": { ... }
}
```

### 错误响应示例

```json
{
    "code": 20001,
    "message": "参数错误",
    "error": "清单 ID 不能为空"
}
```

## 风险评估

* **低风险**：修改范围仅限于 `responser.go` 文件中的一个函数

* **兼容性**：前端代码需要根据 `Code` 字段判断是否成功，而不是依赖 HTTP 状态码

## 验证方法

1. 启动服务并测试 API 接口
2. 使用 curl 或浏览器开发者工具检查响应状态码
3. 验证错误响应的 HTTP 状态码为 200，且响应体中包含正确的错误信息

