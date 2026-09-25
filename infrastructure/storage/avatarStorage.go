package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	userApp "naotodoserver/application/user"
	"naotodoserver/conf"
)

// avatarFilenamePattern 头像文件名白名单：仅允许 {数字}.jpg/.jpeg/.png，
// 同时杜绝路径穿越（..、/、\ 等字符均无法匹配）。
var avatarFilenamePattern = regexp.MustCompile(`^[0-9]{1,20}\.(jpg|jpeg|png)$`)

// isSafeAvatarFilename 判断文件名是否为合法的头像文件名
// @param filename 目标文件名
// @return bool 是否合法
func isSafeAvatarFilename(filename string) bool {
	return avatarFilenamePattern.MatchString(filename)
}

// avatarStorageImpl 头像存储实现（本地文件系统）
type avatarStorageImpl struct{}

// NewAvatarStorage 创建头像存储实例
// @return userApp.AvatarStorage 头像存储实例
func NewAvatarStorage() userApp.AvatarStorage {
	return &avatarStorageImpl{}
}

// avatarDir 获取头像存储目录
// @return string 头像存储目录
func avatarDir() string {
	return filepath.Join(conf.Conf.Uploads.UploadDir, conf.Conf.Uploads.AvatarDir)
}

// Save 保存头像文件
// @param ctx 上下文
// @param filename 目标文件名
// @param src 文件内容
// @return string 可访问的头像相对路径
// @return error 错误
func (a *avatarStorageImpl) Save(
	_ context.Context,
	filename string,
	src io.Reader,
) (string, error) {
	// 1. 确保上传目录存在
	uploadDir := avatarDir()
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		return "", errors.New("创建上传目录失败 - " + err.Error())
	}
	// 2. 写入目标文件
	savePath := filepath.Join(uploadDir, filename)
	dst, err := os.Create(savePath)
	if err != nil {
		return "", errors.New("文件保存失败 - " + err.Error())
	}
	defer dst.Close()
	if _, err = io.Copy(dst, src); err != nil {
		return "", errors.New("文件保存失败 - " + err.Error())
	}
	// 3. 构造可访问的相对路径
	staticPath := conf.Conf.Uploads.StaticPath
	if staticPath == "" {
		staticPath = "/static/uploads"
	}
	return fmt.Sprintf(
		"%s/%s/%s",
		staticPath,
		conf.Conf.Uploads.AvatarDir,
		filename,
	), nil
}

// Open 打开已存在的头像文件流
// @param ctx 上下文
// @param filename 目标文件名
// @return io.ReadCloser 头像文件流
// @return error 错误
func (a *avatarStorageImpl) Open(_ context.Context, filename string) (io.ReadCloser, error) {
	// 1. 校验文件名，防止路径穿越与非法访问
	if !isSafeAvatarFilename(filename) {
		return nil, errors.New("非法头像文件名")
	}
	// 2. 打开目标文件
	f, err := os.Open(filepath.Join(avatarDir(), filename))
	if err != nil {
		return nil, errors.New("头像文件打开失败 - " + err.Error())
	}
	return f, nil
}

// Delete 删除头像文件
// @param ctx 上下文
// @param filename 目标文件名
// @return error 错误
func (a *avatarStorageImpl) Delete(_ context.Context, filename string) error {
	err := os.Remove(filepath.Join(avatarDir(), filename))
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}
