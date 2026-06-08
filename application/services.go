package application

import (
	authApp "naotodoserver/application/auth"
	pomodoroApp "naotodoserver/application/pomodoro"
	projectApp "naotodoserver/application/project"
	tagApp "naotodoserver/application/tag"
	taskApp "naotodoserver/application/task"
	userApp "naotodoserver/application/user"
)

type Services struct {
	Auth     authApp.AuthApp
	User     userApp.UserApp
	Task     taskApp.TaskApp
	Project  projectApp.ProjectApp
	Tag      tagApp.TagApp
	Pomodoro pomodoroApp.PomodoroApp
}

var App *Services
