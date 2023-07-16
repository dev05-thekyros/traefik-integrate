package main

import (
	"bytes"
	"fmt"
	"github.com/hungvtc/traefik-integrate/sso-server/config"
	storage2 "github.com/hungvtc/traefik-integrate/sso-server/repository"
	"github.com/hungvtc/traefik-integrate/sso-server/service/go-kontrol"
	"github.com/hungvtc/traefik-integrate/sso-server/transport"
	"github.com/hungvtc/traefik-integrate/sso-server/wrapper"
	"os"
	"strings"
	"time"

	echoLog "github.com/labstack/gommon/log"
	"github.com/neko-neko/echo-logrus/v2/log"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

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
	err := viper.ReadConfig(bytes.NewBuffer([]byte(config.ConfigDefault)))
	if err != nil {
		logger.Fatal(err)
	}
	viper.SetConfigName("config") // name of config file (without extension)
	viper.AddConfigPath(".")      // optionally look for config in the working directory
	err = viper.ReadInConfig()    // Find and read the config file
	if err != nil {               // Handle errors reading the config file
		logger.Warn(fmt.Sprintf("Fail to read file, use default configure, error detail %v /n", err))
	}
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "__"))
	viper.AutomaticEnv()
	var cfg *config.Config
	err = viper.Unmarshal(&cfg)
	if err != nil {
		logger.Fatal(err)
	}
	//DB
	gormdb, err := storage2.ConnectMySQL(cfg.MySQL)
	if err != nil {
		logger.Fatal(err)
	}

	// storage
	storage := storage2.NewGormStorage()

	// kontrol
	storagekontrol := storage2.NewKontrolStorage()
	kontrol := gokontrol.NewBasicKontrol(storagekontrol)

	ser := &wrapper.Service{
		Logger:         logger,
		Config:         cfg,
		DB:             gormdb,
		Storage:        storage,
		Kontrol:        kontrol,
		StorageKontrol: storagekontrol,
	}

	e := transport.NewEcho(ser)

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
