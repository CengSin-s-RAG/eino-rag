package client

import (
	"agent.article.fp/config"
	"agent.article.fp/util"
	"fmt"
	"github.com/redis/go-redis/v9"
	"go.temporal.io/sdk/client"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
)

var (
	Temporal      client.Client
	SyncTemporal  client.Client
	Redis         *redis.Client
	IvankaContent *gorm.DB
)

func InitPostgres(cfg *config.PostgresConfig) *gorm.DB {
	// 生成gorm链接配置
	if cfg == nil {
		panic("postgres config is nil")
	}

	// 构建 PostgreSQL 连接字符串
	sslmode := cfg.SSLMode
	if sslmode == "" {
		sslmode = "disable"
	}
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.User,
		cfg.Password,
		cfg.DbName,
		sslmode,
	)

	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             0,           // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      false,       // Don't include params in the SQL log
			Colorful:                  false,       // Disable color
		},
	)

	// 初始化 PostgreSQL 连接
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		panic(err)
	}

	// 注册 pgvector 扩展
	if err := db.Exec("CREATE EXTENSION IF NOT EXISTS vector").Error; err != nil {
		log.Printf("Warning: failed to create vector extension: %v", err)
	}

	return db
}

func InitRedis(cfg *config.RedisConfig) {
	if cfg == nil {
		panic("temporal config is nil")
	}

	Redis = redis.NewClient(&redis.Options{
		Addr:     cfg.Addr,
		Password: cfg.Password, // no password set
		DB:       cfg.DB,       // use default DB
	})
}

func Init() {
	util.SystemPrompt = util.InitPrompt(config.Cfg.RagPrompt)
	util.RerankPrompt = util.InitPrompt(config.Cfg.Rerank.Prompt)
	util.QueryRewritePrompt = util.InitPrompt(config.Cfg.Rewrite.Prompt)
	InitRedis(config.Cfg.Redis)
	//client.InitTemporal(config.Cfg.Temporal, &client.Temporal)
	//client.InitTemporal(config.Cfg.SyncTemporal, &client.SyncTemporal)
	IvankaContent = InitPostgres(config.Cfg.Postgres)
	InitMcpClient(config.Cfg.McpServer)
	InitTools()
}

func Close() {
	//Temporal.Close()
	//SyncTemporal.Close()
	McpClient.Close()
	Redis.Close()
}
