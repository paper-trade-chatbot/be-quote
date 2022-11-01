package quoteDao

import (
	"context"

	"github.com/go-redis/redis/v9"
	"github.com/paper-trade-chatbot/be-common/marshaller"
	"github.com/paper-trade-chatbot/be-quote/cache"
	"github.com/paper-trade-chatbot/be-quote/models/redisModels"
)

const (
	first_field = "quote"
)

type QueryModel struct {
	ProductID uint64 `redis:"product_id"`
}

func Gets(ctx context.Context, rds *cache.RedisInstance, queries []*QueryModel) ([]*redisModels.QuoteModel, error) {

	pipeline := rds.Pipeline()

	models := []*redisModels.QuoteModel{}

	values := []*redis.MapStringStringCmd{}

	for _, q := range queries {
		key := first_field + ":" + marshaller.Marshal(ctx, q, "redis", ":")
		values = append(values, pipeline.HGetAll(ctx, key))
		models = append(models, &redisModels.QuoteModel{ProductID: q.ProductID})
	}

	cmd, err := pipeline.Exec(ctx)
	if err != nil {
		return nil, err
	}
	for _, c := range cmd {
		if c.Err() != nil && c.Err() != redis.Nil {
			return nil, c.Err()
		}
	}

	for i, v := range values {
		models[i].ProductID = queries[i].ProductID
		models[i].Quote = v.Val()
	}

	return models, nil

}

func Updates(ctx context.Context, rds *cache.RedisInstance, models []*redisModels.QuoteModel) error {

	for _, m := range models {
		key := first_field + ":" + marshaller.Marshal(ctx, m, "redis", ":")
		if err := rds.HSet(ctx, key, m.Quote).Err(); err != nil {
			return err
		}
	}
	return nil
}

func Delete(ctx context.Context, rds *cache.RedisInstance, models []*redisModels.QuoteModel) error {
	for _, m := range models {
		key := first_field + ":" + marshaller.Marshal(ctx, m, "redis", ":")

		hashIndexes := []string{}
		for k := range m.Quote {
			hashIndexes = append(hashIndexes, k)
		}
		if err := rds.HDel(ctx, key, hashIndexes...).Err(); err != nil {
			return err
		}
	}
	return nil
}
