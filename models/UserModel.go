package models

type User struct {
	ID       uint   `gorm:"primaryKey" json:"id"`
	Name     string `gorm:"size:255" json:"name"`
	Email    string `gorm:"unique" json:"email"`
	Password string `gorm:"size:255" json:"-"`
}
