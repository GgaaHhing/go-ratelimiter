package ip_ratelimiter

import (
	_ "embed"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"log"
	"net/http"
	"time"
)

type Builder struct {
	biz       string
	cmd       redis.Cmdable
	tInterval time.Duration
	rate      int
}

//go:embed slide_window.lua
var luaScript string

func NewBuilder(cmd redis.Cmdable, tInterval time.Duration, rate int) *Builder {
	return &Builder{
		biz:       "ip-rateLimiter",
		cmd:       cmd,
		tInterval: tInterval,
		rate:      rate,
	}
}

func (b *Builder) Build() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ok, err := b.limit(ctx)
		if err != nil || ok {
			log.Println(err)
			log.Println(ok)
			// 为了防止系统崩溃，直接限流
			ctx.AbortWithStatus(http.StatusInternalServerError)
			return
		}

	}
}

func (b *Builder) limit(ctx *gin.Context) (bool, error) {
	// 	我们要拿到IP
	key := fmt.Sprintf("%s:%s", b.biz, ctx.ClientIP())
	return b.cmd.Eval(ctx, luaScript, []string{key},
		b.tInterval, b.rate, time.Now().UnixMilli()).Bool()
}
