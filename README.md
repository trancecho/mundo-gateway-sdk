# mundo-gateway-sdk
网关sdk，public

已实现restful，grpc服务一键注册

todo: 跳过接口希望可以打印跳过了谁。

## 快速上手

# Gateway SDK 快速上手指南

Gateway SDK 是一个用于与网关服务交互的工具包，支持服务注册、心跳信号发送、自动路由注册等功能。以下是快速上手的步骤和示例。

---

## 1. 安装 SDK
确保项目中已经导入了 SDK 包：

```go
import "github.com/trancecho/mundo-gateway-sdk"
```

---

## 2. 初始化 SDK
使用 `NewGatewaySDK` 函数初始化 SDK，传入网关的地址。

```go
gatewayURL := "http://your-gateway-url"
sdk := sdk.NewGatewaySDK(gatewayURL)
```

---

## 3. 注册服务地址
在服务启动时，调用 `RegisterServiceAddress` 方法将服务地址注册到网关。

```go
sdk.ServiceName = "user-service"
sdk.Address = "127.0.0.1:8080"
sdk.Protocol = "http"
sdk.RegisterServiceAddress()
```

---

## 4. 发送心跳信号
定期调用 `SendAliveSignal` 方法，向网关发送心跳信号，表明服务健康。

```go
go func() {
    for {
        sdk.SendAliveSignal("user-service", "127.0.0.1:8080")
        time.Sleep(30 * time.Second) // 每30秒发送一次心跳
    }
}()
```

---

## 5. 自动注册 HTTP 路由
如果使用 Gin 框架，可以调用 `AutoRegisterGinRoutes` 方法自动注册所有路由。

```go
router := gin.Default()
router.GET("/user/info", func(c *gin.Context) { /* 处理逻辑 */ })
router.POST("/user/create", func(c *gin.Context) { /* 处理逻辑 */ })

err := sdk.AutoRegisterGinRoutes(router, "user-service")
if err != nil {
    log.Fatal("自动注册 Gin 路由失败:", err)
}
```

---

## 6. 自动注册 gRPC 路由
如果使用 gRPC 服务，可以调用 `AutoRegisterGRPCRoutes` 方法自动注册所有 gRPC 路由。

```go
grpcServer := grpc.NewServer()
// 注册 gRPC 服务
// ...

err := sdk.AutoRegisterGRPCRoutes(grpcServer, "user-service")
if err != nil {
    log.Fatal("自动注册 gRPC 路由失败:", err)
}
```

---

## 7. 示例代码
以下是一个完整的示例：

```go
package main

import (
	"log"
	"time"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"your_project_path/sdk"
)

func main() {
	// 初始化 SDK
	gatewayURL := "http://your-gateway-url"
	sdk := sdk.NewGatewaySDK(gatewayURL)

	// 注册服务地址
	sdk.ServiceName = "user-service"
	sdk.Address = "127.0.0.1:8080"
	sdk.Protocol = "http"
	sdk.RegisterServiceAddress()

	// 发送心跳信号
	go func() {
		for {
			sdk.SendAliveSignal("user-service", "127.0.0.1:8080")
			time.Sleep(30 * time.Second)
		}
	}()

	// 注册 HTTP 路由
	router := gin.Default()
	router.GET("/user/info", func(c *gin.Context) { /* 处理逻辑 */ })
	router.POST("/user/create", func(c *gin.Context) { /* 处理逻辑 */ })

	err := sdk.AutoRegisterGinRoutes(router, "user-service")
	if err != nil {
		log.Fatal("自动注册 Gin 路由失败:", err)
	}

	// 注册 gRPC 路由
	grpcServer := grpc.NewServer()
	// 注册 gRPC 服务
	// ...

	err = sdk.AutoRegisterGRPCRoutes(grpcServer, "user-service")
	if err != nil {
		log.Fatal("自动注册 gRPC 路由失败:", err)
	}

	// 启动服务
	log.Fatal(router.Run(":8080"))
}
```

---

## 8. 注意事项
- 确保网关地址正确。
- 心跳信号需要定期发送，建议每 5 秒发送一次。
- 路由注册时，确保路径和方法与后端服务一致。
