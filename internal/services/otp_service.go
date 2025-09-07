package services

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type OTPStore interface {
	Set(ctx context.Context, phone string, otp string, ttl time.Duration) error
	Get(ctx context.Context, phone string) (string, error)
	Delete(ctx context.Context, phone string) error
}

type InMemoryOTPStore struct {
	data map[string]otpEntry
}

type otpEntry struct {
	otp string
	exp time.Time
}

func NewInMemoryOTPStore() *InMemoryOTPStore { return &InMemoryOTPStore{data: map[string]otpEntry{}} }

func (s *InMemoryOTPStore) Set(ctx context.Context, phone, otp string, ttl time.Duration) error {
	s.data[phone] = otpEntry{otp: otp, exp: time.Now().Add(ttl)}
	return nil
}
func (s *InMemoryOTPStore) Get(ctx context.Context, phone string) (string, error) {
	e, ok := s.data[phone]
	if !ok || time.Now().After(e.exp) {
		return "", fmt.Errorf("not found or expired")
	}
	return e.otp, nil
}
func (s *InMemoryOTPStore) Delete(ctx context.Context, phone string) error {
	delete(s.data, phone)
	return nil
}

// Redis implementation
type RedisOTPStore struct {
	client *redis.Client
	prefix string
}

func NewRedisOTPStore(client *redis.Client) *RedisOTPStore {
	return &RedisOTPStore{client: client, prefix: "otp:"}
}
func (r *RedisOTPStore) Set(ctx context.Context, phone, otp string, ttl time.Duration) error {
	return r.client.Set(ctx, r.prefix+phone, otp, ttl).Err()
}
func (r *RedisOTPStore) Get(ctx context.Context, phone string) (string, error) {
	return r.client.Get(ctx, r.prefix+phone).Result()
}
func (r *RedisOTPStore) Delete(ctx context.Context, phone string) error {
	return r.client.Del(ctx, r.prefix+phone).Err()
}

// OTP generation
func GenerateOTP(n int) (string, error) {
	var b [8]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", err
	}
	v := binary.LittleEndian.Uint64(b[:])
	format := fmt.Sprintf("%%0%dd", n)
	return fmt.Sprintf(format, v%uint64(pow10(n))), nil
}

func pow10(n int) int {
	p := 1
	for i := 0; i < n; i++ {
		p *= 10
	}
	return p
}
