package api

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type opaqueTokenStore struct {
	client *redis.Client
}

type opaqueTokenRecord struct {
	UserID   string `json:"userId"`
	Email    string `json:"email,omitempty"`
	Username string `json:"username"`
}

func newOpaqueTokenStore(connectionString string) (*opaqueTokenStore, error) {
	if connectionString == "" {
		return nil, nil
	}
	options, err := redis.ParseURL(connectionString)
	if err != nil {
		return nil, fmt.Errorf("parse Redis connection string: %w", err)
	}
	options.DialTimeout = 5 * time.Second
	options.ReadTimeout = 5 * time.Second
	options.WriteTimeout = 5 * time.Second
	return &opaqueTokenStore{client: redis.NewClient(options)}, nil
}

func (s *opaqueTokenStore) lookup(ctx context.Context, token string) (*opaqueTokenRecord, error) {
	if s == nil || s.client == nil {
		return nil, nil
	}
	digest := sha256.Sum256([]byte(token))
	raw, err := s.client.Get(ctx, "bitrok:cli-token:"+hex.EncodeToString(digest[:])).Result()
	if err == redis.Nil {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var record opaqueTokenRecord
	if err := json.Unmarshal([]byte(raw), &record); err != nil || record.UserID == "" || record.Username == "" {
		return nil, nil
	}
	return &record, nil
}
