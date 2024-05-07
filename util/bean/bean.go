package bean

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

type TokenConfig struct {
	JwtKey string `env:"JWT_KEY" envDefault:"secret"`
}

type RedisCfg struct {
	Addr     string `env:"REDIS_ADDR" envDefault:"127.0.0.1"`
	Password string `env:"REDIS_PASSWORD" envDefault:""`
	DB       int    `env:"REDIS_DB" envDefault:"0"`
	URL      string `env:"REDIS_URL"`
}

type PgDbCfg struct {
	User     string `env:"PG_DB_USER" envDefault:"postgres"`
	Address  string `env:"PG_DB_ADDRESS" envDefault:"127.0.0.1"`
	Password string `env:"PG_DB_PASSWORD"`
	Database string `env:"PG_DB_DATABASE" envDefault:"postgres"`
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
	Sender  string `json:"sender,omitempty"`
}

type Claims struct {
	Email string `json:"email,omitempty"`
	UUID  string `json:"uuid"`
	jwt.RegisteredClaims
}

type WhatsappVerificationClaims struct {
	Email         string `json:"email,omitempty"`
	WhatAppNumber string `json:"whatsapp_number,omitempty"`
	jwt.RegisteredClaims
}

type CookieConfig struct {
	Domain string `env:"DOMAIN"`
	Type   string `env:"TYPE"`
}

type WhatsAppConfig struct {
	Baseurl     string `env:"BASE_URL"`
	PhoneID     string `env:"PHONE_ID"`
	AuthToken   string `env:"WHATSAPP_API_TOKEN"`
	VerifyToken string `env:"VERIFY_TOKEN"`
}

type WhatsappEmail struct {
	Email         string `sql:"email" json:"email,omitempty"`
	WhatAppNumber string `sql:"whatsapp_number" json:"whatsapp_number,omitempty"`
}

type MailConfig struct {
	Host     string `env:"MAIL_HOST"`
	Port     string `env:"MAIL_PORT"`
	Username string `env:"MAIL_USERNAME"`
	Password string `env:"MAIL_PASSWORD"`
	From     string `env:"MAIL_FROM"`
}

type EncryptDecryptConfig struct {
	EncryptionKey string `env:"ENCRYPTION_KEY" envDefault:"secret"`
	DecryptionKey string `env:"DECRYPTION_KEY" envDefault:"secret"`
}
