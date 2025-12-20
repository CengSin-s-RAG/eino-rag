package client

import (
	"agent.article.fp/config"
	"github.com/meguminnnnnnnnn/go-openai"
	"github.com/qdrant/go-client/qdrant"
	"github.com/redis/go-redis/v9"
	"go.temporal.io/sdk/client"
	"gorm.io/gorm"
	"log"
)

var (
	Qdrant       *qdrant.Client
	AI           *openai.Client
	Temporal     client.Client
	SyncTemporal client.Client
	Redis        *redis.Client
	Mysql        *gorm.DB
)

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

func InitTemporal(cfg *config.TemporalConfig, destClient *client.Client) {
	if cfg == nil {
		panic("temporal config is nil")
	}

	c, err := client.Dial(client.Options{
		HostPort:  cfg.HostPort,
		Namespace: cfg.Namespace,
	})
	if err != nil {
		log.Fatalln("connntect to temporal failed, err ", err)
	}
	*destClient = c
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

func Close() {
	Temporal.Close()
	SyncTemporal.Close()
	Qdrant.Close()
	McpClient.Close()
	Redis.Close()
}
