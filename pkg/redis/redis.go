package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

var (
	RedisClient *redis.Client
	ctx         = context.Background()
)

// InitRedis menginisialisasi koneksi Redis
func InitRedis() {
	db, _ := strconv.Atoi(os.Getenv("REDIS_DB"))
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%s", os.Getenv("REDIS_HOST"), os.Getenv("REDIS_PORT")),
		Username: os.Getenv("REDIS_USERNAME"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       db,
	})

	if _, err := RedisClient.Ping(ctx).Result(); err != nil {
		panic("Failed to connect to Redis: " + err.Error())
	}
}

// IsConnected checks if Redis client is connected
func IsConnected() bool {
	if RedisClient == nil {
		return false
	}
	_, err := RedisClient.Ping(ctx).Result()
	return err == nil
}

// Ping performs a health check ping to Redis
func Ping() error {
	if RedisClient == nil {
		return fmt.Errorf("redis client not initialized")
	}
	return RedisClient.Ping(ctx).Err()
}

// ============================================
// STRING OPERATIONS
// ============================================

// Set menyimpan value dengan key dan expiration
func Set(key string, value interface{}, expiration time.Duration) error {
	return RedisClient.Set(ctx, key, value, expiration).Err()
}

// Get mengambil value berdasarkan key
func Get(key string) (string, error) {
	return RedisClient.Get(ctx, key).Result()
}

// GetWithDefault mengambil value, return default jika tidak ada
func GetWithDefault(key string, defaultValue string) string {
	val, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return defaultValue
	}
	return val
}

// Delete menghapus key
func Delete(key string) error {
	return RedisClient.Del(ctx, key).Err()
}

// DeleteMultiple menghapus multiple keys sekaligus
func DeleteMultiple(keys ...string) error {
	return RedisClient.Del(ctx, keys...).Err()
}

// Exists mengecek apakah key ada
func Exists(key string) (bool, error) {
	result, err := RedisClient.Exists(ctx, key).Result()
	return result > 0, err
}

// Expire set expiration time untuk key yang sudah ada
func Expire(key string, expiration time.Duration) error {
	return RedisClient.Expire(ctx, key, expiration).Err()
}

// TTL mendapatkan sisa waktu expiration
func TTL(key string) (time.Duration, error) {
	return RedisClient.TTL(ctx, key).Result()
}

// ============================================
// JSON OPERATIONS
// ============================================

// SetJSON menyimpan object sebagai JSON
func SetJSON(key string, value interface{}, expiration time.Duration) error {
	jsonData, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return RedisClient.Set(ctx, key, jsonData, expiration).Err()
}

// GetJSON mengambil dan unmarshal JSON ke object
func GetJSON(key string, dest interface{}) error {
	val, err := RedisClient.Get(ctx, key).Result()
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(val), dest)
}

// ============================================
// COUNTER OPERATIONS
// ============================================

// Increment menambah nilai counter
func Increment(key string) (int64, error) {
	return RedisClient.Incr(ctx, key).Result()
}

// IncrementBy menambah nilai counter dengan jumlah tertentu
func IncrementBy(key string, value int64) (int64, error) {
	return RedisClient.IncrBy(ctx, key, value).Result()
}

// Decrement mengurangi nilai counter
func Decrement(key string) (int64, error) {
	return RedisClient.Decr(ctx, key).Result()
}

// DecrementBy mengurangi nilai counter dengan jumlah tertentu
func DecrementBy(key string, value int64) (int64, error) {
	return RedisClient.DecrBy(ctx, key, value).Result()
}

// ============================================
// HASH OPERATIONS
// ============================================

// HSet menyimpan field-value di hash
func HSet(key string, field string, value interface{}) error {
	return RedisClient.HSet(ctx, key, field, value).Err()
}

// HGet mengambil value dari field di hash
func HGet(key string, field string) (string, error) {
	return RedisClient.HGet(ctx, key, field).Result()
}

// HGetAll mengambil semua field-value di hash
func HGetAll(key string) (map[string]string, error) {
	return RedisClient.HGetAll(ctx, key).Result()
}

// HMSet menyimpan multiple field-value di hash
func HMSet(key string, values map[string]interface{}) error {
	return RedisClient.HMSet(ctx, key, values).Err()
}

// HDelete menghapus field dari hash
func HDelete(key string, fields ...string) error {
	return RedisClient.HDel(ctx, key, fields...).Err()
}

// HExists mengecek apakah field ada di hash
func HExists(key string, field string) (bool, error) {
	return RedisClient.HExists(ctx, key, field).Result()
}

// ============================================
// LIST OPERATIONS
// ============================================

// LPush menambahkan element ke awal list
func LPush(key string, values ...interface{}) error {
	return RedisClient.LPush(ctx, key, values...).Err()
}

// RPush menambahkan element ke akhir list
func RPush(key string, values ...interface{}) error {
	return RedisClient.RPush(ctx, key, values...).Err()
}

// LPop mengambil dan menghapus element dari awal list
func LPop(key string) (string, error) {
	return RedisClient.LPop(ctx, key).Result()
}

// RPop mengambil dan menghapus element dari akhir list
func RPop(key string) (string, error) {
	return RedisClient.RPop(ctx, key).Result()
}

// LRange mengambil range element dari list
func LRange(key string, start, stop int64) ([]string, error) {
	return RedisClient.LRange(ctx, key, start, stop).Result()
}

// LLen mendapatkan panjang list
func LLen(key string) (int64, error) {
	return RedisClient.LLen(ctx, key).Result()
}

// ============================================
// SET OPERATIONS
// ============================================

// SAdd menambahkan member ke set
func SAdd(key string, members ...interface{}) error {
	return RedisClient.SAdd(ctx, key, members...).Err()
}

// SMembers mengambil semua member dari set
func SMembers(key string) ([]string, error) {
	return RedisClient.SMembers(ctx, key).Result()
}

// SIsMember mengecek apakah member ada di set
func SIsMember(key string, member interface{}) (bool, error) {
	return RedisClient.SIsMember(ctx, key, member).Result()
}

// SRem menghapus member dari set
func SRem(key string, members ...interface{}) error {
	return RedisClient.SRem(ctx, key, members...).Err()
}

// SCard mendapatkan jumlah member di set
func SCard(key string) (int64, error) {
	return RedisClient.SCard(ctx, key).Result()
}

// ============================================
// SORTED SET OPERATIONS
// ============================================

// ZAdd menambahkan member dengan score ke sorted set
func ZAdd(key string, score float64, member interface{}) error {
	return RedisClient.ZAdd(ctx, key, &redis.Z{
		Score:  score,
		Member: member,
	}).Err()
}

// ZRange mengambil range member dari sorted set (ascending)
func ZRange(key string, start, stop int64) ([]string, error) {
	return RedisClient.ZRange(ctx, key, start, stop).Result()
}

// ZRevRange mengambil range member dari sorted set (descending)
func ZRevRange(key string, start, stop int64) ([]string, error) {
	return RedisClient.ZRevRange(ctx, key, start, stop).Result()
}

// ZRangeWithScores mengambil range member beserta score
func ZRangeWithScores(key string, start, stop int64) ([]redis.Z, error) {
	return RedisClient.ZRangeWithScores(ctx, key, start, stop).Result()
}

// ZRem menghapus member dari sorted set
func ZRem(key string, members ...interface{}) error {
	return RedisClient.ZRem(ctx, key, members...).Err()
}

// ZScore mendapatkan score dari member
func ZScore(key string, member string) (float64, error) {
	return RedisClient.ZScore(ctx, key, member).Result()
}

// ============================================
// PATTERN OPERATIONS
// ============================================

// Keys mencari keys berdasarkan pattern
func Keys(pattern string) ([]string, error) {
	return RedisClient.Keys(ctx, pattern).Result()
}

// DeleteByPattern menghapus keys berdasarkan pattern
func DeleteByPattern(pattern string) error {
	keys, err := RedisClient.Keys(ctx, pattern).Result()
	if err != nil {
		return err
	}
	if len(keys) > 0 {
		return RedisClient.Del(ctx, keys...).Err()
	}
	return nil
}

// ============================================
// UTILITY FUNCTIONS WITH KEY PREFIXING
// ============================================

// BuildKey membuat key dengan prefix/folder structure
func BuildKey(parts ...string) string {
	key := ""
	for i, part := range parts {
		if i > 0 {
			key += ":"
		}
		key += part
	}
	return key
}

// ============================================
// SPECIFIC DOMAIN FUNCTIONS (EXAMPLES)
// ============================================

// Token operations
func StoreToken(userId string, token string) error {
	expiration, _ := strconv.Atoi(os.Getenv("REDIS_TOKEN_EXPIRATION"))
	return Set(BuildKey("token", userId), token, time.Duration(expiration)*time.Second)
}

func GetToken(userId string) (string, error) {
	return Get(BuildKey("token", userId))
}

func InvalidateToken(userId string) error {
	return Delete(BuildKey("token", userId))
}

// InvalidateUserTokens is an alias for InvalidateToken
func InvalidateUserTokens(userId string) error {
	return InvalidateToken(userId)
}

// Session operations
func StoreSession(sessionId string, data interface{}, expiration time.Duration) error {
	return SetJSON(BuildKey("session", sessionId), data, expiration)
}

func GetSession(sessionId string, dest interface{}) error {
	return GetJSON(BuildKey("session", sessionId), dest)
}

func DeleteSession(sessionId string) error {
	return Delete(BuildKey("session", sessionId))
}

// Cache operations
func SetCache(key string, data interface{}, expiration time.Duration) error {
	return SetJSON(BuildKey("cache", key), data, expiration)
}

func GetCache(key string, dest interface{}) error {
	return GetJSON(BuildKey("cache", key), dest)
}

func InvalidateCache(key string) error {
	return Delete(BuildKey("cache", key))
}

// Rate limiting
func IncrementRateLimit(userId string, expiration time.Duration) (int64, error) {
	key := BuildKey("ratelimit", userId)
	count, err := Increment(key)
	if err != nil {
		return 0, err
	}
	// Set expiration only on first increment
	if count == 1 {
		Expire(key, expiration)
	}
	return count, nil
}

func GetRateLimit(userId string) (int64, error) {
	key := BuildKey("ratelimit", userId)
	val, err := Get(key)
	if err != nil {
		return 0, err
	}
	return strconv.ParseInt(val, 10, 64)
}

// Lock operations (distributed lock)
func AcquireLock(resource string, expiration time.Duration) (bool, error) {
	key := BuildKey("lock", resource)
	result, err := RedisClient.SetNX(ctx, key, "locked", expiration).Result()
	return result, err
}

func ReleaseLock(resource string) error {
	return Delete(BuildKey("lock", resource))
}

// User online status
func SetUserOnline(userId string, expiration time.Duration) error {
	return Set(BuildKey("online", userId), "1", expiration)
}

func IsUserOnline(userId string) (bool, error) {
	return Exists(BuildKey("online", userId))
}

// FlushAll menghapus semua data (hati-hati!)
func FlushAll() error {
	return RedisClient.FlushAll(ctx).Err()
}
