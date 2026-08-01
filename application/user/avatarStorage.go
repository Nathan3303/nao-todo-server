package user

import (
	"context"
	"io"
)

// AvatarStorage 头像存储端口
type AvatarStorage interface {
	// Save 保存头像文件
	// @param ctx 上下文
	// @param filename 目标文件名
	// @param src 文件内容
	// @return string 可访问的头像相对路径
	// @return error 错误
	Save(ctx context.Context, filename string, src io.Reader) (string, error)

	// Delete 删除头像文件
	// @param ctx 上下文
	// @param filename 目标文件名
	// @return error 错误
	Delete(ctx context.Context, filename string) error
}
