package sdk

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

type Client struct {
	GatewayUrl string
}

func NewClient(gatewayUrl string) *Client {
	return &Client{
		GatewayUrl: gatewayUrl,
	}
}

func (this *Client) GetTarget(servicename string) (string, error) {
	// 这里是获取目标地址的逻辑
	// 例如：根据服务名去网关获取目标地址
	// 这里可以使用http请求去获取
	url := fmt.Sprintf("%s/gateway/service/health?service_name=%s", this.GatewayUrl, servicename)
	resp, err := http.Get(url)
	if err != nil {
		log.Println("http.Get error:", err)
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		log.Println("failed to get target address:", resp.Status)
		return "", err
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println("io.ReadAll error:", err)
		return "", err
	}
	type targetResponse struct {
		ErrCode string `json:"err_code"`
		Message string `json:"message"`
		Data    struct {
			Target string `json:"target"`
		} `json:"data"`
	}
	var res targetResponse
	err = json.Unmarshal(body, &res)
	if err != nil {
		log.Println("json.Unmarshal error:", err)
		return "", err
	}
	target := res.Data.Target
	if !strings.HasPrefix(target, "grpc://") && !strings.HasPrefix(target, "grpcs://") {
		log.Println("target address is invalid")
		return "", err
	}
	target = strings.TrimPrefix(target, "grpc://")
	target = strings.TrimPrefix(target, "grpcs://")
	return target, nil
}
