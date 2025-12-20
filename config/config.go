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

type Config struct {
	Qdrant       *QdrantConfig   `yaml:"qdrant"`
	Temporal     *TemporalConfig `yaml:"temporal"`
	SyncTemporal *TemporalConfig `yaml:"syncTemporal"`
	McpServer    string          `yaml:"mcpServer"`
	Redis        *RedisConfig    `yaml:"redis"`
	Cdc          []Cdc           `yaml:"cdc"`
	Mysql        *MysqlConfig    `yaml:"fpMysql"`
}
