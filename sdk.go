package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"io"
	"log"
	"net/http"
	"strings"
)

type GatewaySDK struct {
	GatewayURL string // 网关的地址
}

func NewGatewaySDK(gatewayURL string) *GatewaySDK {
	return &GatewaySDK{
		GatewayURL: gatewayURL,
	}
}

type RouteInfo struct {
	ServiceName string `json:"service_name"`
	Path        string `json:"path"`
	Method      string `json:"method"`
}

type GrpcApiInfo struct {
	ServiceName string `json:"service_name"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	GrpcService string `json:"grpc_service"`
	GrpcMethod  string `json:"grpc_method"`
}

type Response struct {
	ErrCode int64    `json:"err_code"`
	Message string   `json:"message"`
	Data    []string `json:"data"`
}

func (sdk *GatewaySDK) RegisterRoute(route RouteInfo) error {
	jsonData, err := json.Marshal(route)
	if err != nil {
		return err
	}

	resp, err := http.Post(sdk.GatewayURL+"/gateway/api", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("http.Post error:", err)
		return err
	}
	// 拿到resp的body里的内容
	var response Response
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &response)
	log.Println(response)
	if response.ErrCode == 410100 {
		log.Println(response.ErrCode, "http api已存在，跳过注册")
		return nil
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to register route: %s", resp.Status)
	}

	return nil
}

func (sdk *GatewaySDK) RegisterRoutes(routes []RouteInfo) error {
	for _, route := range routes {
		if err := sdk.RegisterRoute(route); err != nil {
			return err
		}
	}
	return nil
}

// 自动化注册 Gin 路由
func (sdk *GatewaySDK) AutoRegisterGinRoutes(router *gin.Engine, serviceName string) error {
	var routes []RouteInfo
	log.Println(router.Routes())

	// 获取 Gin 的所有路由
	for _, route := range router.Routes() {
		log.Println(route.Path, route.Method)
		routes = append(routes, RouteInfo{
			ServiceName: serviceName,
			Path:        route.Path,
			Method:      route.Method,
		})
	}

	// 批量注册路由
	return sdk.RegisterRoutes(routes)
}

// 自动注册GRPC路由
func (sdk *GatewaySDK) AutoRegisterGRPCRoutes(grpcServer *grpc.Server, serviceName string) error {
	// 获取 gRPC 的所有服务
	serviceInfo := grpcServer.GetServiceInfo()
	var routes []GrpcApiInfo
	for svc, info := range serviceInfo {
		for _, method := range info.Methods {
			httpPath := grpcMethodName2HttpPath(method.Name)
			routes = append(routes, GrpcApiInfo{
				ServiceName: serviceName,
				Path:        "/" + httpPath,
				Method:      "POST",
				GrpcService: svc,
				GrpcMethod:  method.Name,
			})
		}
	}
	log.Println(routes)
	// 批量注册路由
	return sdk.RegisterGRPCRoutes(routes)
}

func grpcMethodName2Snake(methodName string) string {
	// 处理method.Name 从驼峰变成蛇形
	// 例如：SayHello转换为 say_hello
	var res string
	for i, r := range methodName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			res += "_" + strings.ToLower(string(r))
		} else {
			res += strings.ToLower(string(r))
		}
	}
	return res
}

func grpcMethodName2HttpPath(methodName string) string {
	// 处理method.Name 从驼峰变成http路由
	// 例如：SayHello转换为 say/hello
	var res string
	for i, r := range methodName {
		if i > 0 && r >= 'A' && r <= 'Z' {
			res += "/" + strings.ToLower(string(r))
		} else {
			res += strings.ToLower(string(r))
		}
	}
	return res
}

func (sdk *GatewaySDK) RegisterGRPCRoutes(routes []GrpcApiInfo) error {
	for _, route := range routes {
		jsonData, err := json.Marshal(route)
		if err != nil {
			return err
		}

		resp, err := http.Post(sdk.GatewayURL+"/gateway/api", "application/json", bytes.NewBuffer(jsonData))
		if err != nil {
			return err
		}
		// 拿到resp的body里的内容
		var response Response
		body, _ := io.ReadAll(resp.Body)
		json.Unmarshal(body, &response)
		if response.ErrCode == 410100 {
			log.Println(response.ErrCode, "grpc api已存在，跳过注册")
			return nil
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to register GRPC route: %s", resp.Status)
		}
	}
	return nil
}
