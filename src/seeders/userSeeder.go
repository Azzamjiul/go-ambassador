package main

import (
	"ambassador/src/database"
	"ambassador/src/models"

	"github.com/bxcodec/faker/v3"
)

func main() {
	database.Connect()

	ambassador := models.User{
		FirstName:    "admin",
		LastName:     "admin",
		Email:        "admin@mail.com",
		IsAmbassador: false,
	}

	ambassador.SetPassword("admin")
	database.DB.Create(&ambassador)

	for i := 0; i < 30; i++ {
		ambassador := models.User{
			FirstName:    faker.FirstName(),
			LastName:     faker.LastName(),
			Email:        faker.Email(),
			IsAmbassador: true,
		}

		ambassador.SetPassword("1234")
		database.DB.Create(&ambassador)
	}
}
