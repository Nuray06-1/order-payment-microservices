package http

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

var rateCtx = context.Background()

func RateLimiter(
	client *redis.Client,
) gin.HandlerFunc {

	return func(
		c *gin.Context,
	) {

		ip := c.ClientIP()

		key := "rate:" + ip

		limit, _ := strconv.Atoi(
			os.Getenv("RATE_LIMIT"),
		)

		window, _ := strconv.Atoi(
			os.Getenv("RATE_WINDOW"),
		)

		count, _ := client.Incr(
			rateCtx,
			key,
		).Result()

		if count == 1 {

			client.Expire(
				rateCtx,
				key,
				time.Duration(window)*time.Second,
			)
		}

		if count > int64(limit) {

			c.JSON(
				http.StatusTooManyRequests,
				gin.H{
					"error": "rate limit exceeded",
				},
			)

			c.Abort()

			return
		}

		c.Next()
	}
}
