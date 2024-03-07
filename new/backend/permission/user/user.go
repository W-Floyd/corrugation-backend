package user

import (
	"net/mail"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

type User struct {
	gorm.Model
	Name               string
	Email              mail.Address
	SaltedPasswordHash []byte
}

func (u *User) VerifyPassword(p string) bool {
	return bcrypt.CompareHashAndPassword(u.SaltedPasswordHash, []byte(p)) == nil
}

func (u *User) SetPassword(p string) error {
	var err error
	u.SaltedPasswordHash, err = bcrypt.GenerateFromPassword([]byte(p), 20) // Increase this for more difficult password cracking
	return err
}
