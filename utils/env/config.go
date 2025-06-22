package env

import (
	"log"
	"time"

	"github.com/joeshaw/envdecode"
)

type Conf struct {
	DB           ConfDB
	Auth         ConfAuth
	Server       ConfServer
	GoogleAuth   ConfGoogleAuth
	IsProduction bool
}

type ConfServer struct {
	Port         int           `env:"SERVER_PORT,required"`
	Debug        bool          `env:"SERVER_DEBUG,required"`
	Secret       []byte        `env:"SECRET_KEY,required"`
	Domain       string        `env:"DOMAIN,default=localhost"`
	Protocol     string        `env:"PROTOCOL,default=http"`
	TimeoutIdle  time.Duration `env:"SERVER_TIMEOUT_IDLE,required"`
	TimeoutRead  time.Duration `env:"SERVER_TIMEOUT_READ,required"`
	TimeoutWrite time.Duration `env:"SERVER_TIMEOUT_WRITE,required"`
}

type ConfDB struct {
	Host     string `env:"DB_HOST,required"`
	Port     int    `env:"DB_PORT,required"`
	Debug    bool   `env:"DB_DEBUG"`
	DBName   string `env:"DB_NAME,required"`
	Username string `env:"DB_USER,required"`
	Password string `env:"DB_PASS,required"`
}

type ConfAuth struct {
	MaxAge int `env:"AUTH_MAX_AGE,required"`
}

type ConfGoogleAuth struct {
	ID     string `env:"GOOGLE_CLIENT_ID,required"`
	Secret string `env:"GOOGLE_CLIENT_SECRET,required"`
}

func (c *Conf) GetBaseURL() string {
	return c.Server.Protocol + "://" + c.Server.Domain
}

func (c *Conf) GetDomain() string {
	return c.Server.Domain
}

func New() *Conf {
	var c Conf
	if err := envdecode.StrictDecode(&c); err != nil {
		log.Fatalf("Failed to decode: %s", err)
	}

	return &c
}
