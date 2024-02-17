package initializers

import (
	"os"

	"github.com/romiras/go-app-demo/internal/services/drivers"
)

var env string

func GetEnv() string {
	if env != "" {
		return env
	}

	env = os.Getenv("GO_ENV")
	if os.Getenv("GO_ENV") == "" {
		env = "development"
	}
	return env
}

func Dependencies() []interface{} {
	return []interface{}{
		NewViper,
		drivers.NewDB,
	}
}
