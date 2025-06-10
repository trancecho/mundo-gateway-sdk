package i

import (
	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
)

type IGatewayV2 interface {
	GrpcConn(server *grpc.Server)
	HttpConn(router *gin.Engine)

	RegisterServiceAddress() bool
	StartHeartbeat()
	AutoRegisterGinRoutes(router *gin.Engine, serviceName string) error
	AutoRegisterGRPCRoutes(grpcServer *grpc.Server, serviceName string) error
	Ping() string
}
