package initializers

import (
	"log"

	"github.com/spf13/viper"
)

func NewViper() *viper.Viper {
	env := GetEnv()
	v := viper.New()
	v.AddConfigPath("./config")
	v.SetConfigName(env)
	v.AutomaticEnv()
	err := v.ReadInConfig()
	if err != nil {
		log.Fatal(err)
	}
	return v
}
