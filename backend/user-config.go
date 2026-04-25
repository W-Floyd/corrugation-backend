package backend

import "gorm.io/gorm"

// UserConfig stores per-user runtime-overridable settings. Username is empty when auth is disabled.
type UserConfig struct {
	gorm.Model
	Username                   string  `gorm:"uniqueIndex"`
	InfinityTextModel          *string
	InfinityImageModel         *string
	InfinityTextQueryPrefix    *string
	InfinityTextDocumentPrefix *string
}

func loadUserConfig(username string) (UserConfig, error) {
	var cfg UserConfig
	err := db.Where(UserConfig{Username: username}).FirstOrCreate(&cfg).Error
	return cfg, err
}

func saveUserConfig(cfg UserConfig) error {
	return db.Where(UserConfig{Username: cfg.Username}).Assign(cfg).FirstOrCreate(&cfg).Error
}

// effectiveInfinityConfig returns the infinity config for a user, falling back to env defaults for nil fields.
func effectiveInfinityConfig(cfg UserConfig) (text, image, queryPrefix, docPrefix string) {
	text = infinityTextModel
	image = infinityImageModel
	queryPrefix = infinityTextQueryPrefix
	docPrefix = infinityTextDocumentPrefix

	if cfg.InfinityTextModel != nil {
		text = *cfg.InfinityTextModel
	}
	if cfg.InfinityImageModel != nil {
		image = *cfg.InfinityImageModel
	}
	if cfg.InfinityTextQueryPrefix != nil {
		queryPrefix = *cfg.InfinityTextQueryPrefix
	}
	if cfg.InfinityTextDocumentPrefix != nil {
		docPrefix = *cfg.InfinityTextDocumentPrefix
	}
	return
}
