# Gin 表单绑定

一个灵活的 Gin 框架表单绑定中间件，支持从路径、查询参数、请求头和请求体中自动绑定参数，并提供可自定义的响应处理。

## 特性

- **多种函数签名**: 支持三种不同的处理器函数签名
- **全面绑定**: 路径参数、查询参数、请求头和请求体
- **类型安全**: 完全支持所有 Go 基础类型、time.Time、time.Duration 和指针
- **默认值**: 通过结构体标签支持默认值
- **可自定义响应**: 可插拔的响应处理接口
- **验证集成**: 与 Gin 内置验证系统配合使用
- **零依赖**: 仅依赖 Gin 框架

## 安装

```bash
go get github.com/zgs225/gin-form-binding
```

## 快速开始

```go
package main

import (
    "github.com/gin-gonic/gin"
    "github.com/zgs225/gin-form-binding"
)

func main() {
    r := gin.Default()
    
    // 创建表单绑定构建器
    builder := ginbinding.NewBasicFormBindingGinHandlerBuilder(nil, nil)
    
    // 定义你的处理器函数
    handler := func(c *gin.Context, req struct {
        Name  string `json:"name" binding:"required"`
        Email string `json:"email" binding:"required,email"`
        Age   int    `json:"age" binding:"min=18"`
    }) (interface{}, error) {
        // 你的业务逻辑
        return gin.H{
            "message": "Hello " + req.Name,
            "email":   req.Email,
            "age":     req.Age,
        }, nil
    }
    
    // 转换为 Gin 处理器
    ginHandler, err := builder.FormBindingGinHandlerFunc(handler)
    if err != nil {
        panic(err)
    }
    
    r.POST("/user", ginHandler)
    r.Run(":8080")
}
```

## 支持的函数签名

库支持三种不同的函数签名：

### 1. 带结构体参数返回错误的函数
```go
func(c *gin.Context, req struct {
    Name string `json:"name"`
}) error {
    // 处理请求
    return nil
}
```

### 2. 带结构体参数返回数据和错误的函数
```go
func(c *gin.Context, req struct {
    Name string `json:"name"`
}) (interface{}, error) {
    // 处理请求
    return gin.H{"message": "Hello " + req.Name}, nil
}
```

### 3. 不带结构体参数返回数据和错误的函数
```go
func(c *gin.Context) (interface{}, error) {
    // 不绑定参数处理请求
    return gin.H{"message": "Hello World"}, nil
}
```

## 支持的标签

### 路径参数
```go
type Request struct {
    ID   int    `path:"id"`
    Name string `path:"name"`
}
```

### 查询参数
```go
type Request struct {
    Page     int    `form:"page"`
    PageSize int    `form:"page_size"`
    Search   string `form:"search"`
}
```

### 请求头
```go
type Request struct {
    UserAgent string `header:"User-Agent"`
    AuthToken string `header:"Authorization"`
}
```

### 请求体
```go
type Request struct {
    Name  string `json:"name"`
    Email string `json:"email"`
    Age   int    `json:"age"`
}
```

### 默认值
```go
type Request struct {
    Name     string        `json:"name" default:"John"`
    Age      int           `json:"age" default:"25"`
    IsActive bool          `json:"is_active" default:"true"`
    Timeout  time.Duration `json:"timeout" default:"30s"`
}
```

## 支持的数据类型

- **字符串**: `string`
- **整数**: `int`, `int8`, `int16`, `int32`, `int64`
- **无符号整数**: `uint`, `uint8`, `uint16`, `uint32`, `uint64`
- **浮点数**: `float32`, `float64`
- **布尔值**: `bool` (支持: true/false, 1/0, yes/no, on/off)
- **时间**: `time.Time` (支持多种格式)
- **持续时间**: `time.Duration`
- **指针**: 所有类型都可以是指针 (`*string`, `*int`, 等)

## 自定义响应处理器

你可以通过实现 `ResponseHandler` 接口来提供自定义响应处理：

```go
type CustomResponseHandler struct{}

func (h *CustomResponseHandler) HandleSuccess(ctx *gin.Context, data interface{}) {
    ctx.JSON(http.StatusOK, gin.H{
        "success": true,
        "data":    data,
    })
}

func (h *CustomResponseHandler) HandleError(ctx *gin.Context, err error) {
    ctx.JSON(http.StatusBadRequest, gin.H{
        "success": false,
        "error":   err.Error(),
    })
}

// 使用自定义处理器
builder := ginbinding.NewBasicFormBindingGinHandlerBuilder(nil, &CustomResponseHandler{})
```

## 验证集成

库与 Gin 的验证系统无缝配合：

```go
import "github.com/gin-gonic/gin/binding"

// 使用 Gin 的默认验证器
builder := ginbinding.NewBasicFormBindingGinHandlerBuilder(binding.Validator, nil)

// 带验证标签的处理器
handler := func(c *gin.Context, req struct {
    Name  string `json:"name" binding:"required"`
    Email string `json:"email" binding:"required,email"`
    Age   int    `json:"age" binding:"min=18,max=100"`
}) (interface{}, error) {
    return gin.H{"message": "Valid data"}, nil
}
```

## 高级示例

### 混合绑定 (路径 + 查询 + 请求头 + 请求体)
```go
handler := func(c *gin.Context, req struct {
    // 路径参数
    UserID int `path:"user_id"`
    
    // 查询参数
    Page int `form:"page"`
    
    // 请求头
    AuthToken string `header:"Authorization"`
    
    // 请求体
    Name  string `json:"name"`
    Email string `json:"email"`
}) (interface{}, error) {
    return gin.H{
        "user_id":    req.UserID,
        "page":       req.Page,
        "auth_token": req.AuthToken,
        "name":       req.Name,
        "email":      req.Email,
    }, nil
}
```

### 嵌入结构体
```go
type Pagination struct {
    Page     int `form:"page" default:"1"`
    PageSize int `form:"page_size" default:"10"`
}

type SearchRequest struct {
    Pagination
    Query string `form:"query"`
}

handler := func(c *gin.Context, req SearchRequest) (interface{}, error) {
    return gin.H{
        "page":      req.Page,
        "page_size": req.PageSize,
        "query":     req.Query,
    }, nil
}
```

### 混合类型的默认值
```go
handler := func(c *gin.Context, req struct {
    Name     string        `json:"name" default:"John"`
    Age      int           `json:"age" default:"25"`
    IsActive bool          `json:"is_active" default:"true"`
    Timeout  time.Duration `json:"timeout" default:"30s"`
    Created  time.Time     `json:"created" default:"2023-01-01T00:00:00Z"`
}) (interface{}, error) {
    return gin.H{
        "name":     req.Name,
        "age":      req.Age,
        "is_active": req.IsActive,
        "timeout":  req.Timeout.String(),
        "created":  req.Created.Format(time.RFC3339),
    }, nil
}
```

## 错误处理

库提供全面的错误处理：

- **绑定错误**: 无效的类型转换、缺少必需字段
- **验证错误**: 自定义验证失败
- **处理器错误**: 从处理器函数返回的错误

所有错误都会传递给你的 `ResponseHandler` 进行自定义格式化。

## API 参考

### 类型

```go
// ResponseHandler 定义处理 HTTP 响应的接口
type ResponseHandler interface {
    HandleSuccess(ctx *gin.Context, data interface{})
    HandleError(ctx *gin.Context, err error)
}

// FormBindingGinHandlerBuilder 定义创建 Gin 处理器的接口
type FormBindingGinHandlerBuilder interface {
    FormBindingGinHandlerFunc(i any) (gin.HandlerFunc, error)
}
```

### 函数

```go
// NewBasicFormBindingGinHandlerBuilder 创建新的构建器
func NewBasicFormBindingGinHandlerBuilder(
    validator binding.StructValidator,
    responseHandler ResponseHandler,
) *BasicFormBindingGinHandlerBuilder

// NewDefaultResponseHandler 创建默认响应处理器
func NewDefaultResponseHandler() *DefaultResponseHandler
```

## 贡献

1. Fork 仓库
2. 创建你的功能分支 (`git checkout -b feature/amazing-feature`)
3. 提交你的更改 (`git commit -m 'Add some amazing feature'`)
4. 推送到分支 (`git push origin feature/amazing-feature`)
5. 打开 Pull Request

## 许可证

本项目采用 MIT 许可证 - 查看 [LICENSE](LICENSE) 文件了解详情。

## 更新日志

### v1.0.0
- 初始版本
- 支持所有三种函数签名
- 全面的类型支持
- 默认值处理
- 自定义响应处理器
- 验证集成
