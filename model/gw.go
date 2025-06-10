package model

type RouteInfo struct {
	ServiceName string `json:"service_name"`
	Path        string `json:"path"`
	Method      string `json:"method"`
}
