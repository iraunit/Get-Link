package util

import (
	"github.com/golang-jwt/jwt/v5"
	"github.com/gorilla/websocket"
	"sync"
)

type MainCfg struct {
	Port string `env:"PORT" envDefault:"1025"`
}

type MiddlewareCfg struct {
	JwtKey   string `env:"JWT_KEY" envDefault:"secret"`
	CorsHost string `env:"CORS_HOST" envDefault:"*"`
}

type RedisCfg struct {
	Addr     string `env:"REDIS_ADDR" envDefault:"127.0.0.1"`
	Password string `env:"REDIS_PASSWORD" envDefault:""`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
	URL      string `env:"REDIS_URL"`
}

type PgDbCfg struct {
	User     string `env:"PG_DB_USER"`
	Address  string `env:"PG_DB_ADDRESS"`
	Password string `env:"PG_DB_PASSWORD"`
	Database string `env:"PG_DB_DATABASE"`
}

type User struct {
	Lock        *sync.Mutex
	Connections []*websocket.Conn
}

type Response struct {
	StatusCode int         `json:"code,omitempty"`
	Result     interface{} `json:"result,omitempty"`
	Error      string      `json:"error,omitempty"`
	Message    string      `json:"message,omitempty"`
}

type GetLink struct {
	ID       int    `sql:"id" json:"id,omitempty"`
	Sender   string `sql:"sender" json:"sender,omitempty"`
	Receiver string `sql:"receiver" json:"receiver,omitempty"`
	Message  string `sql:"message" json:"message,omitempty"`
	UUID     string `sql:"uuid" json:"uuid,omitempty"`
}

type PubSubMessage struct {
	Message string `json:"message,omitempty"`
	UUID    string `json:"uuid,omitempty"`
	ID      int    `json:"id,omitempty"`
}

type Claims struct {
	Email string `json:"email,omitempty"`
	UUID  string `json:"uuid"`
	jwt.RegisteredClaims
}

type CookieConfig struct {
	Domain string `env:"DOMAIN"`
	Type   string `env:"TYPE"`
}
