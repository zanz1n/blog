package xhttp

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/zanz1n/blog/internal/utils"
)

func LogRequest(start time.Time, c *Ctx, err error) {
	const msg = "HTTP: Request"

	req := fmt.Sprintf("%s %s %d", c.Method, c.URL.Path, c.GetStatusCode())

	userId := "nil"
	token, _ := c.GetAuth()
	if token != nil {
		userId = token.ID.String()
	}

	if err != nil {
		slog.Info(msg,
			"error", err,
			"user_id", userId,
			"req", req,
			utils.TookAttr(start, time.Microsecond),
		)
	} else {
		slog.Info(msg,
			"user_id", userId,
			"req", req,
			utils.TookAttr(start, time.Microsecond),
		)
	}
}
