package config


type Config struct {
	Server struct {
		Port int    `yaml:"port"`
		Mode string `yaml:"mode"`
	} `yaml:"server"`

	ClickHouse struct {
		Host     string `yaml:"host"`
		Port     int    `yaml:"port"`
		User     string `yaml:"user"`
		Password string `yaml:"password"`
		Name     string `yaml:"name"`
	} `yaml:"ClickHouse"`

	RabbitMQ struct {
		URL    string   `yaml:"url"`
		Queues []string `yaml:"queues"`
	} `yaml:"RabbitMQ"`

	CreditSrv struct {
		BaseUrl string `yaml:"baseUrl"`
	} `yaml:"CreditSrv"`

	Price struct {
		Normal  int `yaml:"normal"`
		Express int `yaml:"express"`
	} `yaml:"Price"`
}
