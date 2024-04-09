package goal

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/valyala/fasthttp/fasthttpadaptor"
)

type FiberConfig struct {
	Next      func(c *fiber.Ctx) bool
	VisitorID func(c *fiber.Ctx) string
}

var FiberConfigDefault = FiberConfig{
	Next:      nil,
	VisitorID: nil,
}

func fiberConfigDefault(config ...FiberConfig) FiberConfig {
	if len(config) < 1 {
		return FiberConfigDefault
	}
	return config[0]
}

func NewFiberMiddleware(config ...FiberConfig) fiber.Handler {
	return func(c *fiber.Ctx) error {
		cfg := fiberConfigDefault(config...)
		if cfg.Next != nil && cfg.Next(c) {
			return c.Next()
		}
		extra := RecordExtra{
			HitID: GenHitID(),
		}
		if cfg.VisitorID != nil {
			extra.VisitorID = cfg.VisitorID(c)
		}
		c.Locals(Config.GoalIDKey, extra.HitID)
		now := time.Now()
		if err := c.Next(); err != nil {
			return err
		}
		fmt.Println("fibermiddleware path", c.Path())
		var req http.Request
		err := fasthttpadaptor.ConvertRequest(c.Context(), &req, true)
		if err != nil {
			return nil
		}
		extra.ResponseTime = time.Since(now).Milliseconds()
		Record(&req, extra)
		return nil
	}
}
