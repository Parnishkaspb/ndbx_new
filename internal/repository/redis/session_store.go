package redis

import (
	"context"
	"strconv"
	"time"

	goredis "github.com/redis/go-redis/v9"
)

const (
	sessionKeyPrefix = "sid:"
	updatedAtField   = "updated_at"
)

var createSessionScript = goredis.NewScript(`
if redis.call("EXISTS", KEYS[1]) == 1 then
  return 0
end
redis.call("HSET", KEYS[1], "created_at", ARGV[1], "updated_at", ARGV[2])
redis.call("EXPIRE", KEYS[1], ARGV[3])
return 1
`)

// SessionStore implements session persistence in Redis.
type SessionStore struct {
	cli *goredis.Client
}

// NewSessionStore creates redis-backed session store.
func NewSessionStore(cli *goredis.Client) *SessionStore {
	return &SessionStore{cli: cli}
}

// Exists checks if session key exists.
func (s *SessionStore) Exists(ctx context.Context, sid string) (bool, error) {
	n, err := s.cli.Exists(ctx, sessionKey(sid)).Result()
	if err != nil {
		return false, err
	}
	return n > 0, nil
}

// CreateIfNotExists atomically creates session hash and sets TTL.
func (s *SessionStore) CreateIfNotExists(
	ctx context.Context,
	sid string,
	createdAt time.Time,
	updatedAt time.Time,
	ttl time.Duration,
) (bool, error) {
	created, err := createSessionScript.Run(
		ctx,
		s.cli,
		[]string{sessionKey(sid)},
		createdAt.UTC().Format(time.RFC3339),
		updatedAt.UTC().Format(time.RFC3339),
		strconv.Itoa(int(ttl.Seconds())),
	).Int64()
	if err != nil {
		return false, err
	}
	return created == 1, nil
}

// Refresh updates updated_at and extends session TTL.
func (s *SessionStore) Refresh(ctx context.Context, sid string, updatedAt time.Time, ttl time.Duration) error {
	key := sessionKey(sid)
	pipe := s.cli.TxPipeline()
	pipe.HSet(ctx, key, updatedAtField, updatedAt.UTC().Format(time.RFC3339))
	pipe.Expire(ctx, key, ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func sessionKey(sid string) string {
	return sessionKeyPrefix + sid
}
