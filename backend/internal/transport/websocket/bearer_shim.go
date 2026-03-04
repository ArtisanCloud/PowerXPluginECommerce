package websocket

import (
	"encoding/base64"
	"strings"

	"github.com/gin-gonic/gin"
)

func b64urlDecode(raw string) (string, error) {
	raw = strings.ReplaceAll(raw, "-", "+")
	raw = strings.ReplaceAll(raw, "_", "/")
	switch len(raw) % 4 {
	case 2:
		raw += "=="
	case 3:
		raw += "="
	}
	b, err := base64.StdEncoding.DecodeString(raw)
	if err != nil {
		return "", err
	}
	return string(b), nil
}

// BearerShim promotes query/subprotocol token into Authorization header for WS handshake.
func BearerShim() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.GetHeader("Authorization") == "" {
			if auth := c.Query("authorization"); strings.HasPrefix(strings.ToLower(auth), "bearer ") {
				c.Request.Header.Set("Authorization", auth)
			}

			if c.GetHeader("Authorization") == "" {
				vals := c.Request.Header.Values("Sec-WebSocket-Protocol")
				if len(vals) == 0 {
					if v := c.GetHeader("Sec-WebSocket-Protocol"); v != "" {
						vals = []string{v}
					}
				}
				for _, value := range vals {
					for _, proto := range strings.Split(value, ",") {
						normalized := strings.TrimSpace(proto)
						if !strings.HasPrefix(strings.ToLower(normalized), "bearer.") {
							continue
						}
						raw := normalized[strings.IndexByte(normalized, '.')+1:]
						token, err := b64urlDecode(raw)
						if err != nil || token == "" {
							continue
						}
						c.Request.Header.Set("Authorization", "Bearer "+token)
						c.Writer.Header().Set("Sec-WebSocket-Protocol", normalized)
						goto NEXT
					}
				}
			}
		}

	NEXT:
		c.Next()
	}
}
