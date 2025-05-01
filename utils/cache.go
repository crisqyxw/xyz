// utils/cache.go
package utils

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"log"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/nutsdb/nutsdb"
)

var NutsDB *nutsdb.DB

var Ctx = context.Background()

func InitCache() {
	opt := nutsdb.DefaultOptions
	opt.Dir = "/tmp/nutsdb" // 设置数据库存储路径
	db, err := nutsdb.Open(opt)
	if err != nil {
		log.Fatal(err)
	}
	NutsDB = db
}

// GetCacheKey 生成缓存键
func GetCacheKey(uri string, token string) string {
	hash := sha256.Sum256([]byte(uri + ":" + token))
	return hex.EncodeToString(hash[:])
}

// GetCachedResponse 获取缓存内容
func GetCachedResponse(key string) (string, error) {
	var value string
	err := NutsDB.View(func(tx *nutsdb.Tx) error {
		entry, err := tx.Get("cacheBucket", []byte(key))
		if err != nil {
			return err
		}
		value = string(entry)
		return nil
	})
	return value, err
}

// SetCachedResponse 设置缓存内容并存储时间戳
// SetCachedResponse 设置缓存内容并存储时间戳
// SetCachedResponse 设置缓存内容并存储时间戳

func SetCachedResponse(key string, response string, ttl int) {
	err := NutsDB.Update(func(tx *nutsdb.Tx) error {
		return tx.Put("cacheBucket", []byte(key), []byte(response), uint32(ttl))
	})
	if err != nil {
		log.Printf("Failed to set cache: %s", err)
	}
}

// WithConditionalGet 修改后的函数
func WithConditionalGet(handler gin.HandlerFunc) gin.HandlerFunc {
	return func(c *gin.Context) {
		rawURI := c.Request.URL.Path
		token := c.Request.Header.Get("x-jike-access-token")
		cacheKey := GetCacheKey(rawURI, token)

		cached, err := GetCachedResponse(cacheKey)

		log.Printf("Cache Key: %s | Hit: %v", cacheKey, err == nil)
		log.Printf("Cache error message: %s", err)

		if err == nil {
			lastModified, err := GetCachedResponse(cacheKey + ":last_modified")
			if err == nil {
				_, err := time.Parse("Mon, 02 Jan 2006 15:04:05 GMT", lastModified)
				if err == nil {
					c.Header("Last-Modified", lastModified)
				} else {
					log.Printf("Invalid Last-Modified format: %s", lastModified)
				}
			}

			c.Data(200, "application/json", []byte(cached))
			return
		}

		writer := &responseWriter{body: &bytes.Buffer{}, ResponseWriter: c.Writer}
		c.Writer = writer

		handler(c)

		responseBody := writer.body.String()

		SetCachedResponse(cacheKey, responseBody, 5*60)

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
