package goal

import (
	"net"
	"net/http"
	"strings"
)

var trueClientIP = http.CanonicalHeaderKey("True-Client-IP")
var xForwardedFor = http.CanonicalHeaderKey("X-Forwarded-For")
var xRealIP = http.CanonicalHeaderKey("X-Real-IP")

func parseIP(r *http.Request) string {
	addr := r.RemoteAddr
	i := strings.LastIndex(addr, ":")
	if i == -1 {
		i = len(addr)
	}
	ip := addr[:i]
	ip = strings.NewReplacer("[", "", "]", "").Replace(ip)
	if tcip := r.Header.Get(trueClientIP); tcip != "" {
		ip = tcip
	} else if xrip := r.Header.Get(xRealIP); xrip != "" {
		ip = xrip
	} else if xff := r.Header.Get(xForwardedFor); xff != "" {
		i := strings.Index(xff, ",")
		if i == -1 {
			i = len(xff)
		}
		ip = xff[:i]
	}
	if ip == "" || net.ParseIP(ip) == nil {
		return ""
	}
	return ip
}

func parseBrowser(r *http.Request) string {
	ua := r.UserAgent()
	if strings.Contains(ua, "OPR") || strings.Contains(ua, "OPT") {
		return "Opera"
	} else if strings.Contains(ua, "Chrome") {
		return "Chrome"
	} else if strings.Contains(ua, "Firefox") || strings.Contains(ua, "FxiOS") {
		return "Firefox"
	} else if strings.Contains(ua, "Trident") {
		return "Internet Explorer"
	} else if strings.Contains(ua, "Safari") {
		return "Safari"
	} else if strings.Contains(ua, "Edg") {
		return "Edge"
	} else if strings.Contains(ua, "Vivaldi") {
		return "Vivaldi"
	} else if strings.Contains(ua, "YaBrowser") {
		return "Yandex"
	}
	return ""
}

func parseOS(r *http.Request) string {
	ua := r.UserAgent()
	if strings.Contains(ua, "iPhone") || strings.Contains(ua, "Apple-iPhone") {
		return "iPhone"
	} else if strings.Contains(ua, "Windows Phone") {
		return "Windows Phone"
	} else if strings.Contains(ua, "Android") {
		return "Android"
	} else if strings.Contains(ua, "Windows NT") {
		return "Windows"
	} else if strings.Contains(ua, "CrOS") {
		return "ChromeOS"
	} else if strings.Contains(ua, "Macintosh") {
		return "Mac"
	} else if strings.Contains(ua, "Linux") {
		return "Linux"
	}
	return ""
}
