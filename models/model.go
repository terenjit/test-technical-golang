package models

type PhoneNumber struct {
	ID           uint   `json:"id" gorm:"primary_key"`
	PhoneNumbers string `json:"phone_numbers"`
	Provider     string `json:"provider"`
}
