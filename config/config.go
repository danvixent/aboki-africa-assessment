package config

type BaseConfig struct {
	PaystackAPIKey string          `yaml:"paystack_api_key"`
	Postgres       *PostgresConfig `yaml:"postgres"`
}

type PostgresConfig struct {
	Database string `yaml:"database"`
	Host     string `yaml:"host"`
	Port     string `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	MaxConn  int    `yaml:"max_conn"`
}
