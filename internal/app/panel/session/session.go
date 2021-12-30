package session

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/go-redis/redis/v8"
	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
	"grader/internal/app/panel/model"
	"grader/internal/app/panel/storage"
	"grader/pkg/logger"
	"time"
)

// session.Manager interface implementation
var _ Manager = (*Redis)(nil)

type Redis struct {
	issuer         string
	secretKey      []byte
	tokenLifetime  time.Duration
	redis          *redis.Client
	redisKeyPrefix string
	users          storage.UserRepository
}

func (svc *Redis) LoggerComponent() string {
	return "RedisSession"
}

func NewRedis(r *redis.Client, secretKey string, users storage.UserRepository, opts ...Option) *Redis {
	var (
		defaultTokenLifeTime  = time.Hour
		defaultRedisKeyPrefix = "Session"
	)

	s := &Redis{
		redis:          r,
		secretKey:      []byte(secretKey),
		users:          users,
		tokenLifetime:  defaultTokenLifeTime,
		redisKeyPrefix: defaultRedisKeyPrefix,
	}

	for _, opt := range opts {
		opt(s)
	}

	return s
}

type Session struct {
	StartedAt time.Time `json:"started_at"`
	ExpiresAt time.Time `json:"expires_at"`
	UserID    uuid.UUID `json:"user_id"`
}

type Claims struct {
	User model.User `json:"user"`
	jwt.StandardClaims
}

// Create method of session.Creator implementation
func (svc *Redis) Create(ctx context.Context, u *model.User) (string, error) {
	log := logger.Ctx(ctx)
	log.Debug().Str("user-id", u.ID.String()).Msg("Create")

	id := uuid.New().String()

	now := time.Now()
	exp := now.Add(svc.tokenLifetime)

	claims := &Claims{
		User: *u,
		StandardClaims: jwt.StandardClaims{
			Id:        id,
			NotBefore: now.Unix(),
			ExpiresAt: exp.Unix(),
			Issuer:    svc.issuer,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	strToken, err := token.SignedString(svc.secretKey)
	if err != nil {
		log.Error().Err(err).Send()
		return "", fmt.Errorf("jwt encode: %w", err)
	}

	b, err := json.Marshal(&Session{
		UserID:    u.ID,
		StartedAt: now,
		ExpiresAt: exp,
	})
	if err != nil {
		log.Error().Err(err).Send()
		return "", fmt.Errorf("json encode: %w", err)
	}

	if err := svc.redis.Set(ctx, svc.redisKey(claims), string(b), svc.tokenLifetime).Err(); err != nil {
		log.Error().Err(err).Send()
		return "", fmt.Errorf("redis set: %w", err)
	}

	return strToken, nil
}

// Read method of session.Reader implementation
func (svc *Redis) Read(ctx context.Context, tokenString string) (*model.User, error) {
	log := logger.Ctx(ctx)
	log.Debug().Msg("Read request")

	c := &Claims{}

	token, err := jwt.ParseWithClaims(tokenString, c, func(token *jwt.Token) (interface{}, error) {
		return svc.secretKey, nil
	})

	if err != nil {
		log.Debug().Err(err).Msg("ParseWithClaims failed")
		return nil, ErrInvalidToken
	}

	c, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		log.Debug().Str("token", tokenString).Msg("Invalid token")
		return nil, ErrInvalidToken
	}

	key := svc.redisKey(c)

	ss, err := svc.redis.Get(ctx, key).Result()
	if err != nil {
		if err == redis.Nil {
			log.Debug().Err(err).Msg("Redis session not found")
		} else {
			log.Error().Err(err).Msg("Redis get failed")
		}

		return nil, ErrInvalidToken
	}

	s := &Session{}
	if err := json.Unmarshal([]byte(ss), s); err != nil {
		log.Debug().Err(err).Msg("Unable to unmarshall session")
		// delete malformed json from redis
		_ = svc.redis.Del(ctx, key)
		return nil, ErrInvalidToken
	}

	u, err := svc.users.Read(ctx, s.UserID)
	if err != nil {
		log.Debug().Err(err).Send()
		return nil, ErrInvalidToken
	}

	return u, nil
}

func (svc *Redis) redisKey(c *Claims) string {
	return fmt.Sprintf("%s:%s:%s", svc.redisKeyPrefix, c.User.ID, c.StandardClaims.Id)
}
