package client

import (
	"agent.article.fp/config"
	"github.com/qdrant/go-client/qdrant"
	"github.com/redis/go-redis/v9"
	"go.temporal.io/sdk/client"
	"log"
)

var (
	Qdrant       *qdrant.Client
	Temporal     client.Client
	SyncTemporal client.Client
	Redis        *redis.Client
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
