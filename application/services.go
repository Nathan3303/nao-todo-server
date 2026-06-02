package application

import (
	authApp "naotodoserver/application/auth"
	projectApp "naotodoserver/application/project"
	tagApp "naotodoserver/application/tag"
	taskApp "naotodoserver/application/task"
	userApp "naotodoserver/application/user"
)

type Services struct {
	Auth    authApp.AuthApp
	User    userApp.UserApp
	Task    taskApp.TaskApp
	Project projectApp.ProjectApp
	Tag     tagApp.TagApp
}

var App *Services
