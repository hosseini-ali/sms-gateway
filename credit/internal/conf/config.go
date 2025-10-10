package conf

import "fmt"

type Config struct {
	DB    DB    `yaml:"db"`
	Redis Redis `yaml:"redis"`
}

type DB struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
}

func (d DB) Dsn() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%d/%s?sslmode=disable",
		d.Username,
		d.Password,
		d.Host,
		d.Port,
		d.Database,
	)
}

type Redis struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

func (r Redis) Dsn() string {
	return fmt.Sprintf("%s:%d", r.Host, r.Port)
}
