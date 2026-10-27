package user

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"

	"naotodoserver/conf"
)

// avatarRandomBytes 头像文件名随机部分长度（16 字节 ⇒ 32 位 hex = 128 bit）
const avatarRandomBytes = 16

// newAvatarFilename 生成不可枚举的头像文件名（SEC-07）：{userId}-{32 位随机 hex}.{ext}。
// 保留 userId 前缀便于运维定位，随机部分 128 bit 使枚举不可行。
// @param userId 用户 ID
// @param ext 文件扩展名（含点，如 ".png"）
// @return string 文件名
// @return error 随机源不可用时的错误
func newAvatarFilename(userId int64, ext string) (string, error) {
	buf := make([]byte, avatarRandomBytes)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("读取随机源失败: %w", err)
	}
	return fmt.Sprintf("%d-%s%s", userId, hex.EncodeToString(buf), ext), nil
}

// localAvatarFilename 从用户当前头像字段提取「本服务头像目录内」的文件名。
// 外部 URL、其它路径、含路径分隔符的非法名字一律返回空串（调用方据此跳过清理）。
// @param avatar 用户表的头像字段（本服务为相对路径，如 /static/uploads/avatars/1.png）
// @return string 文件名（如 1.png / 1-<hex>.png）；无法识别时为空串
func localAvatarFilename(avatar string) string {
	staticPath := strings.TrimSuffix(conf.Conf.Uploads.StaticPath, "/")
	if staticPath == "" {
		staticPath = "/static/uploads"
	}
	avatarDir := strings.Trim(conf.Conf.Uploads.AvatarDir, "/")
	if avatarDir == "" {
		avatarDir = "avatars"
	}
	prefix := staticPath + "/" + avatarDir + "/"
	if !strings.HasPrefix(avatar, prefix) {
		return ""
	}
	name := strings.TrimPrefix(avatar, prefix)
	if name == "" || strings.ContainsAny(name, `/\`) {
		return ""
	}
	return name
}
