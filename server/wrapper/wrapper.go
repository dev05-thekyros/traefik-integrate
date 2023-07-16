package wrapper

import (
	"github.com/hungvtc/traefik-integrate/sso-server/config"
	"github.com/hungvtc/traefik-integrate/sso-server/repository"
	"github.com/hungvtc/traefik-integrate/sso-server/service/go-kontrol"
	"github.com/neko-neko/echo-logrus/v2/log"
)

type Service struct {
	Config         *config.Config
	Logger         *log.MyLogger
	DB             repository.Database
	Storage        repository.Storage
	Kontrol        gokontrol.Kontrol
	StorageKontrol gokontrol.KontrolStore
}
