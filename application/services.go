package application

import (
	authApp "naotodoserver/application/auth"
	commentApp "naotodoserver/application/comment"
	eventApp "naotodoserver/application/event"
	projectApp "naotodoserver/application/project"
	tagApp "naotodoserver/application/tag"
	taskApp "naotodoserver/application/task"
	userApp "naotodoserver/application/user"
)

// Services 应用层服务容器，统一管理所有应用服务的单例
type Services struct {
	Auth    authApp.AuthApp
	Event   eventApp.EventApp
	Comment commentApp.CommentApp
	Project projectApp.ProjectApp
	Tag     tagApp.TagApp
	Task    taskApp.TaskApp
	User    userApp.UserApp
}

// App 全局应用服务实例
var App *Services
