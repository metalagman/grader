package session

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"grader/pkg/logger"
	"grader/pkg/token"
	"net/http"
	"time"
)

// session.Manager interface implementation
var _ Manager = (*Redis)(nil)

type Redis struct {
	tokenManager token.Manager
	redis        *redis.Client

	cookieName      string
	redisKeyPrefix  string
	sessionLifetime time.Duration
}

func NewRedis(r *redis.Client, tm token.Manager, opts ...RedisOption) *Redis {
	const (
		defaultRedisKeyPrefix = "Session"
	)

	s := &Redis{
		redis:        r,
		tokenManager: tm,

		sessionLifetime: defaultSessionLifetime,
		redisKeyPrefix:  defaultRedisKeyPrefix,
		cookieName:      defaultCookieName,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

type RedisOption func(*Redis)

func WithRedisKeyPrefix(v string) RedisOption {
	return func(s *Redis) {
		s.redisKeyPrefix = v
	}
}

func WithSessionLifetime(v time.Duration) RedisOption {
	return func(s *Redis) {
		s.sessionLifetime = v
	}
}

func WithCookieName(v string) RedisOption {
	return func(s *Redis) {
		s.cookieName = v
	}
}

// Create method of session.Manager implementation
func (svc *Redis) Create(ctx context.Context, w http.ResponseWriter, id token.Identity) error {
	l := logger.Ctx(ctx)
	uid := id.Identity()
	sid := uuid.New().String()
	l.Debug().Str("user-id", uid).Str("session-id", sid).Msg("Session create")

	now := time.Now()
	exp := now.Add(svc.sessionLifetime)
	s := &Session{
		ID:        sid,
		UserID:    uid,
		StartedAt: now,
		ExpiresAt: exp,
	}

	tk, err := svc.tokenManager.Issue(s, svc.sessionLifetime)
	if err != nil {
		return fmt.Errorf("token issue: %w", err)
	}

	b, err := json.Marshal(s)
	if err != nil {
		return fmt.Errorf("json encode: %w", err)
	}

	if err := svc.redis.Set(ctx, svc.redisKey(s), string(b), svc.sessionLifetime).Err(); err != nil {
		return fmt.Errorf("redis set: %w", err)
	}

	cookie := &http.Cookie{
		Name:    svc.cookieName,
		Value:   tk,
		Expires: time.Now().Add(svc.sessionLifetime),
		Path:    "/",
	}
	http.SetCookie(w, cookie)

	return nil
}

func (svc *Redis) Read(ctx context.Context, r *http.Request) (*Session, error) {
	l := logger.Ctx(ctx)
	l.Debug().Msg("Session read")

	cookie, err := r.Cookie(svc.cookieName)
	if err == http.ErrNoCookie {
		return nil, ErrUnauthorized
	}

	sessionID, err := svc.tokenManager.Decode(cookie.Value)
	if err != nil {
		return nil, fmt.Errorf("token decode: %w", err)
	}

	sessionKey := svc.redisKey(sessionID)

	ss, err := svc.redis.Get(ctx, sessionKey).Result()
	if err != nil {
		if err == redis.Nil {
			l.Debug().Err(err).Msg("Redis session not found")
		} else {
			l.Error().Err(err).Msg("Redis get failed")
		}

		return nil, ErrUnauthorized
	}

	s := &Session{}
	if err := json.Unmarshal([]byte(ss), s); err != nil {
		l.Debug().Err(err).Msg("Unable to unmarshall session")
		// delete malformed json from redis
		_ = svc.redis.Del(ctx, sessionKey)
		return nil, ErrUnauthorized
	}

	return s, nil
}

func (svc *Redis) DestroyCurrent(ctx context.Context, w http.ResponseWriter, r *http.Request) error {
	defer func() {
		cookie := http.Cookie{
			Name:    svc.cookieName,
			Expires: time.Now().AddDate(0, 0, -1),
			Path:    "/",
		}
		http.SetCookie(w, &cookie)
	}()

	s, err := svc.Read(ctx, r)
	if err != nil {
		return fmt.Errorf("current read: %w", err)
	}

	key := svc.redisKey(s)
	_ = svc.redis.Del(ctx, key)

	return nil
}

func (svc *Redis) redisKey(s token.Identity) string {
	return fmt.Sprintf("%s:%s", svc.redisKeyPrefix, s.Identity())
}
