package goal

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"strings"
)

type Hit struct {
	ID               string
	Path             string
	Query            string
	Title            string
	VisitorID        string `db:"visitor_id"`
	IP               string
	Referer          string
	UserAgent        string `db:"user_agent"`
	Browser          string
	OS               string
	ResponseTime     int64 `db:"response_time"` // ms
	TimeOnPage       int64 `db:"time_on_page"`  // ms
	Width            int
	Height           int
	DevicePixelRatio int   `db:"device_pixel_ratio"`
	IsBot            bool  `db:"is_bot"`
	CreatedAt        int64 `db:"created_at"` // ms
}

func GenHitID() string {
	size := 10
	var buf [12]byte
	var b64 string
	for len(b64) < size {
		rand.Read(buf[:])
		b64 = base64.StdEncoding.EncodeToString(buf[:])
		b64 = strings.NewReplacer("+", "", "/", "").Replace(b64)
	}
	return fmt.Sprintf("%s", b64[0:size])
}
