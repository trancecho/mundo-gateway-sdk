package sdk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/trancecho/mundo-gateway-sdk/i"
	"github.com/trancecho/mundo-gateway-sdk/model"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

//

// v2版本，2025年5月

//	type RedisTokenGetter interface {
//		GetToken() (string, error)
//	}
type GatewayService struct {
	ServiceName string
	Address     string
	Protocol    string
	GatewayURL  string // 网关的地址
	TokenGetter *MyRedisTokenGetter
	Password    string
}

func (this *GatewayService) HttpConn(router *gin.Engine) {
	ok := this.RegisterServiceAddress()
	if !ok {
		this.RegisterServiceAddress()
	}
	this.StartHeartbeat()
	// 自动注册Gin路由
	err := this.AutoRegisterGinRoutes(router, this.ServiceName)
	if err != nil {
		log.Println("网关注册api警报", err)
	}
}

func (this *GatewayService) GrpcConn(server *grpc.Server) {
	ok := this.RegisterServiceAddress()
	if !ok {
		this.RegisterServiceAddress()
	}
	this.StartHeartbeat()
	// 内部做反射
	reflection.Register(server)
	err := this.AutoRegisterGRPCRoutes(server, this.ServiceName)
	if err != nil {
		log.Println("网关注册api警报", err)
	}
}

var _ i.IGatewayV2 = &GatewayService{}

func NewGatewayService(serviceName, address, protocol, gatewayURL string, getter *MyRedisTokenGetter) *GatewayService {
	res := &GatewayService{
		ServiceName: serviceName,
		Address:     address,
		Protocol:    protocol,
		GatewayURL:  gatewayURL,
		TokenGetter: getter,
	}
	token, err := res.TokenGetter.GetToken()
	if err != nil {
		log.Println("获取Token失败，请检查redis是否有token:", err)
	} else {
		res.Password = token
	}
	log.Println("GatewayService初始化成功:", res.ServiceName, res.Address, res.Protocol, res.GatewayURL)
	return res
}

func (g *GatewayService) RegisterServiceAddress() bool {
	// 注册服务地址到网关
	url := fmt.Sprintf("%s/gateway/service", g.GatewayURL)
	data := map[string]string{
		"name":     g.ServiceName,
		"prefix":   "/" + g.ServiceName,
		"protocol": g.Protocol,
		"address":  g.Address,
		"password": g.Password,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("json.Marshal error:", err)
		return false
	}

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("http.NewRequest error:", err)
		return false
	}
	req.Header.Set("Content-Type", "application/json")

	//if g.TokenGetter != nil {
	//	token, err := g.TokenGetter.GetToken()
	//	if err == nil && token != "" {
	//		req.Header.Set("Authorization", "Bearer "+token)
	//	}
	//}

	client := &http.Client{}
	resp, err := client.Do(req)

	if err != nil {
		log.Println("http.Post error:", err)
		return false
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var errCodeReceiver ErrorCodeReceiver
		decoder := json.NewDecoder(resp.Body)
		if err := decoder.Decode(&errCodeReceiver); err != nil {
			fmt.Println("解析 JSON 失败:", err)
			return false
		}
		log.Println("errCodeReceiver:", errCodeReceiver)
		if errCodeReceiver.ErrorCode == RedisDynamicPasswordError {
			token, err := g.TokenGetter.GetToken()
			if err != nil {
				log.Println("TokenGetter.GetToken error:", err)
			}
			g.Password = token
		}
		log.Println("failed to register service address:", resp.Status)
		return false
	}
	return true
}

const (
	RedisDynamicPasswordError = "Error.RedisDynamicPassword"
)

type ErrorCodeReceiver struct {
	ErrorCode any `json:"err_code"`
}

func (g *GatewayService) sendAliveSignal(serviceName string, address string) {
	// 发送心跳信号到网关
	url := fmt.Sprintf("%s/gateway/service/beat", g.GatewayURL)
	data := map[string]string{
		"service_name": serviceName,
		"address":      address,
	}
	jsonData, err := json.Marshal(data)
	if err != nil {
		log.Println("json.Marshal error:", err)
		return
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		log.Println("http.Post error:", err)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("failed to send alive signal:", resp.Status)
	}
}

func (g *GatewayService) StartHeartbeat() {
	// 启动心跳信号
	go func() {
		for {
			g.sendAliveSignal(g.ServiceName, g.Address)
			time.Sleep(10 * time.Second)
		}
	}()
}

type GrpcApiInfo struct {
	ServiceName string `json:"service_name"`
	Path        string `json:"path"`
	Method      string `json:"method"`
	GrpcService string `json:"grpc_service"`
	GrpcMethod  string `json:"grpc_method"`
}

type Response struct {
	ErrCode any      `json:"err_code"`
	Message string   `json:"message"`
	Data    []string `json:"data"`
}

func (sdk *GatewayService) registerRoute(route model.RouteInfo) error {
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

func (sdk *GatewayService) registerRoutes(routes []model.RouteInfo) error {
	for _, route := range routes {
		if err := sdk.registerRoute(route); err != nil {
			return err
		}
	}
	return nil
}

// 自动化注册 Gin 路由
func (sdk *GatewayService) AutoRegisterGinRoutes(router *gin.Engine, serviceName string) error {
	var routes []model.RouteInfo
	//log.Println(router.Routes())

	// 获取 Gin 的所有路由
	for _, route := range router.Routes() {
		log.Println(route.Path, route.Method)
		routes = append(routes, model.RouteInfo{
			ServiceName: serviceName,
			Path:        route.Path,
			Method:      route.Method,
		})
	}

	// 批量注册路由
	return sdk.registerRoutes(routes)
}

// 自动注册GRPC路由
func (sdk *GatewayService) AutoRegisterGRPCRoutes(grpcServer *grpc.Server, serviceName string) error {
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
	return sdk.registerGRPCRoutes(routes)
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

func (sdk *GatewayService) registerGRPCRoutes(routes []GrpcApiInfo) error {
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
			continue
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("failed to register GRPC route: %s", resp.Status)
		}
	}
	return nil
}

func (sdk *GatewayService) Ping() string {
	url := fmt.Sprintf("%s/gateway/ping", sdk.GatewayURL)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("Ping error:", err)
		return "Ping failed"
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("Ping failed with status:", resp.Status)
		return "Ping failed with status: " + resp.Status
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("Read body error:", err)
		return "Ping failed to read body"
	}

	return string(body)
}
