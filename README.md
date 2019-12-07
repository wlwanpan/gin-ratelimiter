# gin-ratelimiter

A light weight in-memory rate limiter middleware for gin-gonic.

## Installation

```
go get github.com/wlwanpan/gin-ratelimiter
```

## Usage

```go
import (
  "github.com/wlwanpan/gin-ratelimiter"
  "github.com/gin-gonic/gin"
)

var limiter = ratelimiter.New(5, 10) // 5 request/s, 10 burst limit.

router := gin.Default()

router.Use(limiter.Limit())
```
