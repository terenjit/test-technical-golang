package databases

import (
	"test-technical-golang/models"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func init() {
	var err error
	dsn := "host=localhost port=5432 user=postgres dbname=test sslmode=disable password=test123"
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		panic(err)
	}

	DB.AutoMigrate(&models.PhoneNumber{})
}
