package client

import (
	"agent.article.fp/config"
	"fmt"
	"github.com/qdrant/go-client/qdrant"
	"github.com/redis/go-redis/v9"
	"go.temporal.io/sdk/client"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
)

var (
	Qdrant        *qdrant.Client
	Temporal      client.Client
	SyncTemporal  client.Client
	Redis         *redis.Client
	IvankaContent *gorm.DB
)

func InitMysql(cfg *config.MysqlConfig) *gorm.DB {
	// 生成gorm链接配置
	if cfg == nil {
		panic("mysql config is nil")
	}

	// 构建 DSN 连接字符串
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True",
		cfg.User,
		cfg.Password,
		cfg.Host,
		cfg.Port,
		cfg.DbName,
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

	// 初始化 MySQL 连接
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: newLogger})
	if err != nil {
		panic(err)
	}
	return db
}

func InitQdrant(cfg *config.QdrantConfig) {
	if cfg == nil {
		panic("qdrant config is nil")
	}

	c, err := qdrant.NewClient(&qdrant.Config{
		Host: cfg.Host,
		Port: cfg.Port,
	})
	if err != nil {
		log.Fatalln("qdrant client init failed, err ", err)
	}

	Qdrant = c
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
	InitRedis(config.Cfg.Redis)
	InitQdrant(config.Cfg.Qdrant)
	//client.InitTemporal(config.Cfg.Temporal, &client.Temporal)
	//client.InitTemporal(config.Cfg.SyncTemporal, &client.SyncTemporal)
	IvankaContent = InitMysql(config.Cfg.IvankaContent)
	if len(config.Cfg.MCP) > 0 {
		InitMcpClient(config.Cfg.MCP[0].Endpoint)
		InitTools()
	}
}

func Close() {
	//Temporal.Close()
	//SyncTemporal.Close()
	Qdrant.Close()
	if McpClient != nil {
		McpClient.Close()
	}
	Redis.Close()
}
