package main

import (
	"errors"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

type Map struct {
	X float64 `json:"lat" gorm:"not null"`
	Y float64 `json:"lon" gorm:"not null"`
}

const (
	userMaxAge              = 150
	userMinAge              = 14
	userMaxFirstnameLength  = 255
	userMinFirstnameLength  = 3
	userMaxSecondnameLength = 255
	userMinSecondnameLength = 3
	userMinRandomKeyLength  = 3
	userMaxRandomKeyLength  = 8
)

type User struct {
	ID         string `json:"id"`
	Firstname  string `gorm:"not null" json:"firstname"`
	Secondname string `gorm:"not null" json:"secondname"`
	Age        int    `gorm:"not null" json:"age"`
	RandomKey  string `gorm:"not null"`
	Map        Map    `gorm:"embedded" json:"map"`
}

type DbErrorKind int

const (
	None DbErrorKind = iota
	NotFoundError
	DbError
)

func (user *User) Validate() (bool, string) {
	if len(user.Firstname) < userMinFirstnameLength || len(user.Firstname) > userMaxFirstnameLength {
		return false, "firstname invalid"
	}

	if len(user.Secondname) < userMinSecondnameLength || len(user.Secondname) > userMaxSecondnameLength {
		return false, "secondname invalid"
	}

	if user.Age < userMinAge || user.Age > userMaxAge {
		return false, "invalid age"
	}

	if len(user.RandomKey) < userMinRandomKeyLength || len(user.RandomKey) > userMaxRandomKeyLength {
		return false, "invalid random key"
	}

	return true, ""
}

func (user *User) Create() DbErrorKind {
	result := db.Create(user)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return NotFoundError
		} else {
			return DbError
		}
	}

	return None
}

func (user *User) GetById() DbErrorKind {
	err := db.First(user, "id = ?", user.ID).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFoundError
		} else {
			return DbError
		}
	}

	return None
}

func (user *User) Update() DbErrorKind {
	err := db.Save(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFoundError
		} else {
			return DbError
		}
	}

	return None
}

func (user *User) UpdateSelective() DbErrorKind {
	err := db.Model(user).Updates(user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return NotFoundError
		} else {
			return DbError
		}
	}

	return None
}

func (user *User) ToDisplay() UserDisplayDTO {
	return UserDisplayDTO{ID: user.ID, Firstname: user.Firstname, Secondname: user.Secondname, Age: user.Age, Map: user.Map}
}

type UserCreationDTO struct {
	Firstname  string `gorm:"not null" json:"firstname" validate:"required"`
	Secondname string `gorm:"not null" json:"secondname" validate:"required"`
	Age        int    `gorm:"not null" json:"age" validate:"required,gte=14,lte=150"`
	Map        Map    `gorm:"embedded" json:"map" validate:"required"`
}

func (user UserCreationDTO) Validate() (bool, string) {
	if (len(user.Firstname) < userMinFirstnameLength || len(user.Firstname) > userMaxFirstnameLength) && len(user.Firstname) != 0 {
		return false, "firstname invalid"
	}

	if (len(user.Secondname) < userMinSecondnameLength || len(user.Secondname) > userMaxSecondnameLength) && len(user.Secondname) != 0 {
		return false, "secondname invalid"
	}

	if (user.Age < userMinAge || user.Age > userMaxAge) && user.Age != 0 {
		return false, "invalid age"
	}

	return true, ""
}

func (user UserCreationDTO) CombineWithUser(userToCombine User) User {
	res := userToCombine
	res.Firstname = user.Firstname
	res.Secondname = user.Secondname
	res.Age = user.Age
	res.Map = user.Map

	return res
}

func (user UserCreationDTO) ToUser() User {
	return User{Firstname: user.Firstname, Secondname: user.Secondname, Age: user.Age, Map: user.Map}
}

type UserDisplayDTO struct {
	ID         string `json:"id"`
	Firstname  string `gorm:"not null" json:"firstname"`
	Secondname string `gorm:"not null" json:"secondname"`
	Age        int    `gorm:"not null" json:"age"`
	Map        Map    `gorm:"embedded" json:"map"`
}

func (user UserDisplayDTO) Validate() (bool, string) {
	if len(user.Firstname) < userMinFirstnameLength || len(user.Firstname) > userMaxFirstnameLength {
		return false, "firstname invalid"
	}

	if len(user.Secondname) < userMinSecondnameLength || len(user.Secondname) > userMaxSecondnameLength {
		return false, "secondname invalid"
	}

	if user.Age < userMinAge || user.Age > userMaxAge {
		return false, "invalid age"
	}

	return true, ""
}

func (user UserDisplayDTO) ValidateButIgnoreZeroed() (bool, string) {
	if (len(user.Firstname) < userMinFirstnameLength || len(user.Firstname) > userMaxFirstnameLength) && len(user.Firstname) != 0 {
		return false, "firstname invalid"
	}

	if (len(user.Secondname) < userMinSecondnameLength || len(user.Secondname) > userMaxSecondnameLength) && len(user.Secondname) != 0 {
		return false, "secondname invalid"
	}

	if (user.Age < userMinAge || user.Age > userMaxAge) && user.Age != 0 {
		return false, "invalid age"
	}

	return true, ""
}

func (user UserDisplayDTO) ToUser() User {
	return User{ID: user.ID, Firstname: user.Firstname, Secondname: user.Secondname, Age: user.Age, Map: user.Map}
}

func (user UserDisplayDTO) CombineWithUser(userToCombine User) User {
	res := userToCombine
	res.Firstname = user.Firstname
	res.Secondname = user.Secondname
	res.Age = user.Age
	res.Map = user.Map

	return res
}

var db *gorm.DB

func initializeStorage() {
	createFileIfDoesntExist("storage.db")

	_db, err := gorm.Open(sqlite.Open("storage.db"), &gorm.Config{})
	if err != nil {
		panic("failed to connect database")
	}

	db = _db

	db.AutoMigrate(&User{})


}
