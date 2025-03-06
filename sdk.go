package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
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

func (sdk *GatewaySDK) RegisterRoute(route RouteInfo) error {
	jsonData, err := json.Marshal(route)
	if err != nil {
		return err
	}

	resp, err := http.Post(sdk.GatewayURL+"/gateway/api", "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return err
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
