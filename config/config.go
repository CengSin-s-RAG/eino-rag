package config

type TemporalConfig struct {
	HostPort  string `yaml:"hostPort"`
	Namespace string `yaml:"namespace"`
}

type RedisConfig struct {
	Addr     string `yaml:"addr"`
	Password string `yaml:"password"`
	DB       int    `yaml:"DB"`
}

type Cdc struct {
	Addr      string   `yaml:"addr" json:"addr"`
	User      string   `yaml:"user" json:"user"`
	Password  string   `yaml:"password" json:"password"`
	DbName    string   `yaml:"dbName" json:"dbName"`
	TableName []string `yaml:"tableName" json:"tableName"`
}

type PostgresConfig struct {
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	DbName   string `yaml:"dbname"`
	SSLMode  string `yaml:"sslmode"`
}

type ContextProcessModelConfig struct {
	Prompt string `yaml:"prompt"`
	Url    string `yaml:"url"`
}

type Config struct {
	Postgres      *PostgresConfig            `yaml:"postgres"`
	Temporal      *TemporalConfig            `yaml:"temporal"`
	SyncTemporal  *TemporalConfig            `yaml:"syncTemporal"`
	McpServer     string                     `yaml:"mcpServer"`
	Redis         *RedisConfig               `yaml:"redis"`
	Cdc           []Cdc                      `yaml:"cdc"`
	RagPrompt     string                     `yaml:"ragPromptPath"`
	Rerank        *ContextProcessModelConfig `yaml:"rerank"`
	Rewrite       *ContextProcessModelConfig `yaml:"rewrite"`
}

var (
	Cfg *Config
)
