package handler

import (
	"database/sql"
	"forum_backend/repository"
	"forum_backend/service"
)

type Application struct {
	UserService *service.UserService
}

func InitApp(database *sql.DB) *Application {
	userRepo := &repository.UserRepo{DB: database}
	userService := &service.UserService{UserRepo: userRepo}

	//COMING SOON: services for LIKE,POST,COMMENT,CATEGORY

	return &Application{UserService: userService}
}
