package controllers

import (
	"gorm.io/gorm"
)

type BaseController struct {
	DB *gorm.DB
}

// Constructor function
func NewBaseController(db *gorm.DB) *BaseController {
	return &BaseController{DB: db}
}
