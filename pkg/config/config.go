package config

import (
	"strings"

	"github.com/spf13/viper"
)

const (
	KeyAddr         = "server.addr"
	KeySessionFile  = "game.session_file"
	KeyInitialScore = "game.initial_score"
)

func init() {
	viper.SetDefault(KeyAddr, ":8787")
	viper.SetDefault(KeySessionFile, "session.json")
	viper.SetDefault(KeyInitialScore, 25000)

	viper.SetEnvPrefix("MSO")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
}

func Init() {
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.mahjongsoul-overlay")

	viper.ReadInConfig()
}

func Addr() string        { return viper.GetString(KeyAddr) }
func SessionFile() string { return viper.GetString(KeySessionFile) }
func InitialScore() int   { return viper.GetInt(KeyInitialScore) }
