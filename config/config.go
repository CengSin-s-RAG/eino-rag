package config

type QdrantConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

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

type MysqlConfig struct {
	Host     string `yaml:"host"`
	Port     int64  `yaml:"port"`
	User     string `yaml:"userName"`
	Password string `yaml:"password"`
	DbName   string `yaml:"DB"`
}

type ContextProcessModelConfig struct {
	Prompt string `yaml:"prompt"`
	Url    string `yaml:"url"`
}

type ExecConfig struct {
	AllowedCommands []string `yaml:"allowedCommands"` // 白名单命令列表
	Timeout         int      `yaml:"timeout"`         // 超时时间（秒），默认 60
}

type Config struct {
	Qdrant        *QdrantConfig              `yaml:"qdrant"`
	Temporal      *TemporalConfig            `yaml:"temporal"`
	SyncTemporal  *TemporalConfig            `yaml:"syncTemporal"`
	McpServer     string                     `yaml:"mcpServer"`
	Redis         *RedisConfig               `yaml:"redis"`
	Cdc           []Cdc                      `yaml:"cdc"`
	IvankaContent *MysqlConfig               `yaml:"ivankaContent"`
	RagPrompt     string                     `yaml:"ragPromptPath"`
	SkillsPath    string                     `yaml:"skillsPath"`
	Exec          *ExecConfig                `yaml:"exec"`
	Rerank        *ContextProcessModelConfig `yaml:"rerank"`
	Rewrite       *ContextProcessModelConfig `yaml:"rewrite"`
}

var (
	Cfg *Config
)
