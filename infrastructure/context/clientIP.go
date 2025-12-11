package context

import "context"

type key string

var clientIPKey key = "clientIP"

type ClientInfo struct {
	IP4        string
	IP6        string
	IPRegion   string
	DeviceType string
}

func GetClientInfo(ctx context.Context) ClientInfo {
	clientInfo, ok := ctx.Value(clientIPKey).(ClientInfo)
	if !ok {
		return ClientInfo{}
	}
	return clientInfo
}

func SetClientInfo(ctx context.Context, clientInfo ClientInfo) context.Context {
	return context.WithValue(ctx, clientIPKey, clientInfo)
}
