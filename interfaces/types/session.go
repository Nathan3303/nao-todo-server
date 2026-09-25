package types

// SessionRes 会话列表项响应
type SessionRes struct {
	Id         string `json:"id"`
	DeviceId   string `json:"deviceId"`
	DeviceType string `json:"deviceType"`
	IP4        string `json:"ip4"`
	Region     string `json:"region"`
	CreatedAt  string `json:"createdAt"`
	UpdatedAt  string `json:"updatedAt"`
	Current    bool   `json:"current"`
}
