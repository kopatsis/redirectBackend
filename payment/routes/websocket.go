package routes

import (
	"context"
	"net/http"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/stripe/stripe-go/v72"
	"github.com/stripe/stripe-go/v72/sub"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func WebSocketHandler(rdb *redis.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		id := c.Param("id")

		conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		defer conn.Close()

		pubsub := rdb.Subscribe(context.Background(), "Subscription")
		defer pubsub.Close()

		timeoutDuration := 180 * time.Second
		ticker := time.NewTicker(9 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case msg := <-pubsub.Channel():
				parts := strings.Split(msg.Payload, " --- ")
				if len(parts) == 2 {
					subscriptionID, status := parts[0], parts[1]
					if subscriptionID == id {
						if status == "Success" {
							conn.WriteMessage(websocket.TextMessage, []byte("refresh"))
						} else if status == "Fail" {
							conn.WriteMessage(websocket.TextMessage, []byte("error"))
						}
					}
				}
			case <-ticker.C:
				s, err := sub.Get(id, nil)
				if err == nil && (s.Status == stripe.SubscriptionStatusActive || s.Status == stripe.SubscriptionStatusPastDue) {
					conn.WriteMessage(websocket.TextMessage, []byte("refresh"))
				}
			case <-time.After(timeoutDuration - time.Since(start)):
				conn.WriteMessage(websocket.TextMessage, []byte("timeout"))
				return
			}
		}
	}
}
