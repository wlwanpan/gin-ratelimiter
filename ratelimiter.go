package ratelimiter

import (
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/karlseguin/ccache"
	"golang.org/x/time/rate"
)

const DEFAULT_CACHE_TTL = time.Minute * 5

var RateErr = errors.New("too many requests")

// Limiter is a wrapper around the go rate limiter package
// to handle rate limiting on a request context level.
type Limiter struct {
	sync.Mutex
	reqCache   *ccache.Cache
	RateLimit  rate.Limit
	BucketSize int
}

func New(limit float64, size int) *Limiter {
	return &Limiter{
		Mutex:      sync.Mutex{},
		reqCache:   ccache.New(ccache.Configure()),
		RateLimit:  rate.Limit(limit),
		BucketSize: size,
	}
}

func (gl *Limiter) getReqLimiter(rid string) *rate.Limiter {
	item := gl.reqCache.Get(rid)
	if item != nil {
		return item.Value().(*rate.Limiter)
	}

	// First time request hit from ip, create a new limiter and add it
	// to request cache.
	l := rate.NewLimiter(gl.RateLimit, gl.BucketSize)
	gl.reqCache.Set(rid, l, DEFAULT_CACHE_TTL)
	return l
}

// Wrapper around the time/rate Allow func that first retrives the
// limiter for a given request id (ip address) from the cache.
func (gl *Limiter) isAllow(rid string) bool {
	gl.Lock()
	l := gl.getReqLimiter(rid)
	gl.Unlock()

	return l.Allow()
}

// RateLimit gin middleware that limit the requests per second per context.
// Each request is identified by the ip address and cached in the limiter,
// until dropped by the lru cache or until expired.
func (gl *Limiter) Limit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ipAddr := extractIPAddr(c)

		if !gl.isAllow(ipAddr) {
			c.AbortWithError(http.StatusTooManyRequests, RateErr)
			return
		}
		c.Next()
	}
}

func extractIPAddr(c *gin.Context) string {
	ip, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err != nil {
		// TODO: figure out a fallback in cases where ip parsing from remote addr fails.
		return ""
	}
	return ip
}
