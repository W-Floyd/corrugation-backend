package backend

import "gorm.io/gorm"

// User stores user identity and per-user runtime-overridable settings. Username is empty when auth is disabled.
type User struct {
	gorm.Model
	Username                   string  `gorm:"uniqueIndex"`
	InfinityTextModel          *string
	InfinityImageModel         *string
	InfinityTextQueryPrefix    *string
	InfinityTextDocumentPrefix *string
}

func userUsername(u *User) *string {
	if u == nil {
		return nil
	}
	return &u.Username
}

func loadUser(username string) (User, error) {
	var u User
	err := db.Where(User{Username: username}).FirstOrCreate(&u).Error
	return u, err
}

func saveUser(u User) error {
	return db.Where(User{Username: u.Username}).Assign(u).FirstOrCreate(&u).Error
}

// effectiveInfinityConfig returns the infinity config for a user, falling back to env defaults for nil fields.
func effectiveInfinityConfig(u User) (text, image, queryPrefix, docPrefix string) {
	text = infinityTextModel
	image = infinityImageModel
	queryPrefix = infinityTextQueryPrefix
	docPrefix = infinityTextDocumentPrefix

	if u.InfinityTextModel != nil {
		text = *u.InfinityTextModel
	}
	if u.InfinityImageModel != nil {
		image = *u.InfinityImageModel
	}
	if u.InfinityTextQueryPrefix != nil {
		queryPrefix = *u.InfinityTextQueryPrefix
	}
	if u.InfinityTextDocumentPrefix != nil {
		docPrefix = *u.InfinityTextDocumentPrefix
	}
	return
}
