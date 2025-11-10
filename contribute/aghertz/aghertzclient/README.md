# aghertzclient

Hertz 客户端基础包，为 aggo 代码生成提供统一的客户端调用基础。

## 概述

`aghertzclient` 是一个基于 CloudWeGo Hertz 框架的通用客户端基础包，提供了统一的 HTTP 客户端调用接口，支持服务发现、序列化、请求构建等核心功能。

## 功能特性

- ✅ **统一客户端接口**: 提供标准化的 HTTP 客户端调用方式
- ✅ **服务发现支持**: 支持基于服务名的服务发现机制
- ✅ **灵活序列化**: 支持 JSON、Protobuf 等多种序列化格式
- ✅ **路径参数处理**: 自动处理 RESTful 风格的路径参数
- ✅ **查询参数支持**: 便捷的查询参数设置
- ✅ **内容类型配置**: 支持自定义请求内容类型
- ✅ **错误处理**: 完善的错误处理和日志记录
- ✅ **配置选项**: 丰富的客户端配置选项

## 安装

```bash
go get ag-core/contribute/aghertz/aghertzclient
```

## 快速开始

### 1. 创建基础客户端

```go
package main

import (
    "context"
    "ag-core/contribute/aghertz/aghertzclient"
    "github.com/cloudwego/hertz/pkg/app/client"
)

func main() {
    // 创建 Hertz 客户端
    hc, _ := client.NewClient()
    
    // 创建基础客户端
    baseClient := aghertzclient.NewHertzBaseClient(hc,
        aghertzclient.WithDirectEndpoint("http://localhost:8080"),
    )
    
    // 使用客户端...
}
```

### 2. 使用服务发现

```go
// 使用服务发现
baseClient := aghertzclient.NewHertzBaseClient(hc,
    aghertzclient.WithSDEndpoint("service-name"),
)

// 或者使用直接端点
baseClient := aghertzclient.NewHertzBaseClient(hc,
    aghertzclient.WithDirectEndpoint("http://192.168.1.100:8080"),
)
```

## 核心组件

### HertzBaseClient

基础客户端结构体，提供统一的请求执行接口。

```go
type HertzBaseClient struct {
    endpoint string
    isSD     bool
    cli      *client.Client
}
```

### RequestParam

请求参数配置结构体。

```go
type RequestParam struct {
    Method      string            // HTTP 方法 (GET, POST, PUT, DELETE 等)
    Path        string            // 请求路径，支持路径参数 (:param)
    PathVars    map[string]string // 路径参数映射
    QueryParams map[string]string // 查询参数
    ContentType string            // 内容类型
    Serializer  Serializer        // 序列化器
}
```

### 序列化器

支持多种序列化格式：

- **JSON**: `application/json`
- **Protobuf**: `application/x-protobuf`

```go
// 注册自定义序列化器
aghertzclient.RegisterSerializer("application/xml", &XMLSerializer{})
```

## 使用示例

### 基本 GET 请求

```go
func ExampleGetRequest() {
    var resp MyResponse
    
    reqParam := &aghertzclient.RequestParam{
        Method: "GET",
        Path:   "/api/v1/users/:id",
        PathVars: map[string]string{
            "id": "123",
        },
        QueryParams: map[string]string{
            "fields": "name,email",
        },
    }
    
    err := baseClient.DoRequest(context.Background(), reqParam, nil, &resp)
    if err != nil {
        // 处理错误
    }
}
```

### POST 请求带请求体

```go
func ExamplePostRequest() {
    req := &MyRequest{
        Name:  "John",
        Email: "john@example.com",
    }
    var resp MyResponse
    
    reqParam := &aghertzclient.RequestParam{
        Method:      "POST",
        Path:        "/api/v1/users",
        ContentType: "application/json",
    }
    
    err := baseClient.DoRequest(context.Background(), reqParam, req, &resp)
    if err != nil {
        // 处理错误
    }
}
```

### 使用 Protobuf 序列化

```go
func ExampleProtobufRequest() {
    req := &pb.UserRequest{
        Id: 123,
    }
    var resp pb.UserResponse
    
    reqParam := &aghertzclient.RequestParam{
        Method:      "POST",
        Path:        "/api/v1/user",
        ContentType: "application/x-protobuf",
    }
    
    err := baseClient.DoRequest(context.Background(), reqParam, req, &resp)
    if err != nil {
        // 处理错误
    }
}
```

## 代码生成示例

`aghertzclient` 主要用于代码生成工具生成具体的服务客户端。以下是一个生成的客户端示例：

```go
// 生成的客户端接口
type UserServiceHertzClient interface {
    GetUser(ctx context.Context, req *pb.GetUserRequest, opts ...config.RequestOption) (resp *pb.GetUserResponse, err error)
    CreateUser(ctx context.Context, req *pb.CreateUserRequest, opts ...config.RequestOption) (resp *pb.CreateUserResponse, err error)
}

// 生成的客户端实现
type UserServiceHertzClientImpl struct {
    *aghertzclient.HertzBaseClient
}

func (c *UserServiceHertzClientImpl) GetUser(ctx context.Context, req *pb.GetUserRequest, opts ...config.RequestOption) (*pb.GetUserResponse, error) {
    var resp pb.GetUserResponse
    
    reqParam := &aghertzclient.RequestParam{
        Method: "GET",
        Path:   "/v1/users/:id",
        PathVars: map[string]string{
            "id": req.GetId(),
        },
    }
    
    err := c.DoRequest(ctx, reqParam, nil, &resp, opts...)
    if err != nil {
        return nil, err
    }
    
    return &resp, nil
}
```

## 配置选项

### WithSD

启用服务发现模式。

```go
baseClient := aghertzclient.NewHertzBaseClient(hc,
    aghertzclient.WithSD(true),
)
```

### WithEndpoint

设置端点（服务名或直接地址）。

```go
baseClient := aghertzclient.NewHertzBaseClient(hc,
    aghertzclient.WithEndpoint("user-service"),
)
```

### WithSDEndpoint

设置服务发现端点。

```go
baseClient := aghertzclient.NewHertzBaseClient(hc,
    aghertzclient.WithSDEndpoint("user-service"),
)
```

### WithDirectEndpoint

设置直接端点（不经过服务发现）。

```go
baseClient := aghertzclient.NewHertzBaseClient(hc,
    aghertzclient.WithDirectEndpoint("http://localhost:8080"),
)
```

## 错误处理

客户端提供了完善的错误处理机制：

- HTTP 状态码非 200 时会返回错误
- 序列化/反序列化失败时会返回错误
- 网络请求失败时会返回错误并记录日志

```go
err := baseClient.DoRequest(ctx, reqParam, req, &resp)
if err != nil {
    // 处理不同类型的错误
    if strings.Contains(err.Error(), "status") {
        // HTTP 状态错误
    } else if strings.Contains(err.Error(), "marshal") {
        // 序列化错误
    } else {
        // 网络或其他错误
    }
}
```

## 日志记录

客户端使用 `slog` 进行日志记录：

- 错误请求会记录错误日志
- 成功请求会记录调试日志
- 包含请求方法、URL、状态码等信息

## 扩展性

### 自定义序列化器

```go
type XMLSerializer struct{}

func (s *XMLSerializer) Marshal(v interface{}) ([]byte, error) {
    // 实现 XML 序列化逻辑
    return xml.Marshal(v)
}

func (s *XMLSerializer) ContentType() string {
    return "application/xml"
}

// 注册自定义序列化器
aghertzclient.RegisterSerializer("application/xml", &XMLSerializer{})
```

### 自定义配置选项

```go
func WithTimeout(timeout time.Duration) aghertzclient.AgHertzClientOption {
    return func(c *aghertzclient.HertzBaseClient) {
        // 实现自定义配置
    }
}
```

## 依赖

- `github.com/cloudwego/hertz`: Hertz HTTP 框架
- `github.com/bytedance/sonic`: 高性能 JSON 序列化
- `google.golang.org/protobuf`: Protobuf 支持

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
