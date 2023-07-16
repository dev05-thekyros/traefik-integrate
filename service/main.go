package main

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"time"

	echoLog "github.com/labstack/gommon/log"
	"github.com/neko-neko/echo-logrus/v2/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

type Service struct {
	Config  *Config
	Logger  *log.MyLogger
	DB      Database
	Storage Storage
}

func main() {
	// Logger
	logger := log.Logger()
	logger.SetOutput(os.Stdout)
	logger.SetLevel(echoLog.DEBUG)
	logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})

	// Configuration
	viper.SetConfigType("yaml")
	err := viper.ReadConfig(bytes.NewBuffer([]byte(ConfigDefault)))
	if err != nil {
		logger.Fatal(err)
	}
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	err = viper.ReadInConfig()    // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		logger.Warn(fmt.Sprintf("Fail to read file, use default configure, error detail %v /n", err))
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))
	viper.AutomaticEnv()
	var cfg *Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		logger.Fatal(err)
	}
	//DB
	gormdb, err := ConnectMySQL(cfg.MySQL)
	if err != nil {
		logger.Fatal(err)
	}

	ser := &Service{
		Logger:  logger,
		Config:  cfg,
		DB:      gormdb,
		Storage: NewGormStorage,
	}

	e := NewEcho(ser)

	switch cfg.Environment {
	case "dev":
		err = e.Start(":" + cfg.HTTPPort)
	case "test":
	default:
		err = e.Start(":" + cfg.HTTPPort)
	}
	if err != nil {
		logger.Fatal(err)
	}
}
