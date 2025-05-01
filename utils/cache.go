// utils/cache.go
package utils

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var RedisClient *redis.Client
var Ctx = context.Background()

func InitCache() {
	RedisClient = redis.NewClient(&redis.Options{
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
	})
}

// GetCacheKey 生成缓存键
func GetCacheKey(uri string, token string, bodyHash string) string {
	hash := sha256.Sum256([]byte(uri + ":" + token + ":" + bodyHash))
	return hex.EncodeToString(hash[:])
}

// GetCachedResponse 获取缓存内容
func GetCachedResponse(key string) (string, error) {
	return RedisClient.Get(Ctx, key).Result()
}

// SetCachedResponse 设置缓存内容并存储时间戳
// SetCachedResponse 设置缓存内容并存储时间戳
// SetCachedResponse 设置缓存内容并存储时间戳
func SetCachedResponse(key string, response string, ttl int) {
	now := time.Now().UTC().Format("Mon, 02 Jan 2006 15:04:05 GMT") // 强制使用 GMT 格式
	RedisClient.Set(Ctx, key, response, 0)
	RedisClient.Set(Ctx, key+":last_modified", now, 0)
}

// WithConditionalGet 修改后的函数
func WithConditionalGet(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {

		var bodyBytes []byte
		if c.Request.Body != nil {
			bodyBytes, _ = c.GetRawData()
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes)) // 恢复请求体
		}

		// 计算请求体的哈希值
		bodyHash := ""
		if len(bodyBytes) > 0 {
			hash := sha256.Sum256(bodyBytes)
			bodyHash = hex.EncodeToString(hash[:])
		}
		// 构造缓存键
		rawURI := c.Request.URL.Path
		token := c.Request.Header.Get("x-jike-access-token")
		cacheKey := GetCacheKey(rawURI, token, bodyHash)

		// 获取缓存值
		_, err := GetCachedResponse(cacheKey)

		// 如果启用调试可记录日志
		log.Printf("Cache Key: %s | Hit: %v", cacheKey, err == nil)

		// 缓存存在，添加 Last-Modified Header 并返回缓存的响应数据
		if err == nil {
			// 获取缓存的存储时间
			lastModified, err := RedisClient.Get(Ctx, cacheKey+":last_modified").Result()
			log.Printf("Last-Modified: %s", lastModified)
			if err == nil {
				// 验证时间格式是否符合 RFC1123 或 GMT
				_, err := time.Parse("Mon, 02 Jan 2006 15:04:05 GMT", lastModified)
				if err == nil {
					c.Header("Last-Modified", lastModified)
				} else {
					log.Printf("Invalid Last-Modified format: %s", lastModified)
				}
			}
		}

		// 缓存未命中，继续执行 handler 并捕获响应
		writer := &responseWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = writer

		handler(c)

		responseBody := writer.body.String()

		// 更新缓存
		SetCachedResponse(cacheKey, responseBody, 5*60) // TTL 5分钟
	}
}

type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(data []byte) (int, error) {
	w.body.Write(data)
	return w.ResponseWriter.Write(data)
}
