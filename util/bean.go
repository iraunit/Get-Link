package util

type MainCfg struct {
	Port string `env:"PORT" envDefault:"1025"`
}
