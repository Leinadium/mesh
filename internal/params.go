package internal

import "github.com/caarlos0/env/v11"

type Params struct {
	Port         int      `env:"PORT" envDefault:"8080"`
	DatabaseDSN  string   `env:"DATABASE_DSN"`
	AllowOrigins []string `env:"ALLOW_ORIGINS"`
}

func NewParams() (p Params, err error) {
	err = env.Parse(&p)
	return
}
