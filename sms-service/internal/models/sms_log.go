package models

import "time"

type SMSLog struct {
	Message     string    `json:"message"`
	PhoneNumber string    `json:"phone_number"`
	Org         string    `json:"org"`
	IsExpress   bool      `json:"is_express"`
	CreatedAt   time.Time `json:"created_at"`
}
