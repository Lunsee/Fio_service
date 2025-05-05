package models

import (
	"time"
)

type Persons struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	Name       string    `json:"name" gorm:"column:name"`
	Surname    string    `json:"surname" gorm:"column:surname"`
	Patronymic string    `json:"patronymic" gorm:"column:patronymic"`
	Age        int       `json:"age" gorm:"column:age"`
	Gender     string    `json:"gender" gorm:"column:gender"`
	Ethnicity  string    `json:"ethnicity" gorm:"column:ethnicity"`
	CreatedAt  time.Time `json:"createdAt" gorm:"column:created_at;autoCreateTime"`
	UpdatedAt  time.Time `json:"updatedAt" gorm:"column:updated_at;autoUpdateTime"`
}
