package realtime

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/whg517/node-pulse/pulse/internal/auth"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 70 * time.Second
	pingPeriod     = 30 * time.Second
	maxMessageSize = 1024
)

type inboundMessage struct {
	Type string `json:"type"`
}

// Handler exposes the websocket endpoint for authenticated operator UI clients.
type Handler struct {
	hub        *Hub
	jwtService *auth.JWTService
	upgrader   websocket.Upgrader
}

// NewHandler creates a realtime websocket handler.
func NewHandler(hub *Hub, jwtService *auth.JWTService) *Handler {
	return &Handler{
		hub:        hub,
		jwtService: jwtService,
		upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
	}
}

// ServeWS upgrades an authenticated request to a websocket connection.
func (h *Handler) ServeWS(c *gin.Context) {
	if h.hub == nil || h.jwtService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"code":    "ERR_REALTIME_UNAVAILABLE",
			"message": "Realtime event stream is unavailable",
		})
		return
	}

	claims, ok := h.authenticate(c)
	if !ok {
		return
	}

	conn, err := h.upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}

	client := h.hub.Register(claims.UserID, claims.Role)
	defer h.hub.Unregister(client)

	ctx, cancel := context.WithCancel(c.Request.Context())
	defer cancel()

	go h.writePump(ctx, conn, client)
	h.readPump(ctx, conn, client)
}

func (h *Handler) authenticate(c *gin.Context) (*auth.Claims, bool) {
	token := c.Query("token")
	if token == "" {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "ERR_MISSING_TOKEN",
			"message": "Missing websocket token",
		})
		return nil, false
	}

	claims, err := h.jwtService.ValidateAccessToken(token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "ERR_INVALID_TOKEN",
			"message": "Invalid websocket token",
		})
		return nil, false
	}

	revoked, err := h.jwtService.CheckRevoked(c.Request.Context(), claims.JTI)
	if err != nil || revoked {
		c.JSON(http.StatusUnauthorized, gin.H{
			"code":    "ERR_INVALID_TOKEN",
			"message": "Invalid websocket token",
		})
		return nil, false
	}

	return claims, true
}

func (h *Handler) readPump(ctx context.Context, conn *websocket.Conn, client *Client) {
	defer func() {
		_ = conn.Close()
	}()

	conn.SetReadLimit(maxMessageSize)
	_ = conn.SetReadDeadline(time.Now().Add(pongWait))
	conn.SetPongHandler(func(string) error {
		return conn.SetReadDeadline(time.Now().Add(pongWait))
	})

	for {
		var message inboundMessage
		if err := conn.ReadJSON(&message); err != nil {
			return
		}

		if message.Type == "ping" {
			client.send(ctx, Message{
				Type:      EventPong,
				Payload:   gin.H{"ok": true},
				Timestamp: time.Now().UTC(),
			})
		}
	}
}

func (h *Handler) writePump(ctx context.Context, conn *websocket.Conn, client *Client) {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		_ = conn.Close()
	}()

	for {
		select {
		case message, ok := <-client.sendCh:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				_ = conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}
			if err := conn.WriteJSON(message); err != nil {
				return
			}
		case <-ticker.C:
			_ = conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		case <-ctx.Done():
			return
		}
	}
}
