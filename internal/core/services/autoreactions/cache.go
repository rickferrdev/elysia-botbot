package autoreactions

import (
	"context"
	"encoding/json"
	"time"
)

const autoreactionsCacheTTL = time.Minute * 10

func (service *Service) getCacheKey(guildID, channelID string) string {
	return "autoreactions:" + guildID + ":" + channelID
}

func (service *Service) getAllCacheKey(guildID string) string {
	return "autoreactions:guild:" + guildID
}

func (service *Service) getCache(ctx context.Context, key string, dest any) (bool, error) {
	data, err := service.redis.DB.Get(ctx, key).Bytes()
	if err != nil {
		return false, err
	}

	if err := json.Unmarshal(data, dest); err != nil {
		return false, err
	}

	return true, nil
}

func (service *Service) setCache(ctx context.Context, key string, value any) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}

	return service.redis.DB.Set(
		ctx,
		key,
		data,
		autoreactionsCacheTTL,
	).Err()
}

func (service *Service) delCache(ctx context.Context, guildID, channelID string) error {
	return service.redis.DB.Del(
		ctx,
		service.getCacheKey(guildID, channelID),
		service.getAllCacheKey(guildID),
	).Err()
}
