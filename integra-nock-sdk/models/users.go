package models

import (
	"errors"
	"github.com/jinzhu/gorm"
	"integra-nock-sdk/database"
)

type User struct {
	gorm.Model
	Username string `gorm:"size:255;not null;unique" json:"username"`
	Password string `gorm:"size:255;not null;" json:"password"`
}

//get particular user by user id
func GetUserByID(uid uint) (User,error) {

	var user User

	if err := database.Connector.First(&user,uid).Error; err != nil {
		return user,errors.New("User with this id not found!")
	}

	//user.PrepareGive()
	
	return user,nil
}

func (user *User) PrepareGive(){
	user.Password = ""
}

