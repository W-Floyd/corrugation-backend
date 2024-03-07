package main

import "fmt"

type Config struct {
	DatabaseType     DatabaseType `required:"true" split_words:"true"`
	DatabaseName     string       `required:"true" default:"corrugation" split_words:"true"`
	DatabaseUsername string       `required:"true" split_words:"true"`
	DatabasePassword string       `split_words:"true"`
	DatabaseAddress  string       `required:"true" split_words:"true"`
	DatabaseTimezone string       `required:"true" split_words:"true"`
	Port             int          `default:"8083" split_words:"true"`
	JWTSecret        string       `required:"true" split_words:"true"`
	AssetPath        string       `required:"true" default:"./assets/" split_words:"true"`
}

type DatabaseType string

func (t *DatabaseType) Decode(v string) error {
	for key := range DbHandlers {
		if DatabaseType(v) == key {
			*t = DatabaseType(v)
			return nil
		}
	}
	return fmt.Errorf("Database type '" + v + "' is not known")
}
