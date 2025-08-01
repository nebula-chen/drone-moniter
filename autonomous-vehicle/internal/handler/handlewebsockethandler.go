package handler

import (
	"net/http"

	"autonomous-vehicle/internal/svc"
	"autonomous-vehicle/internal/websocket"
)

func HandleWebSocketHandler(svcCtx *svc.ServiceContext) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := websocket.Upgrader.Upgrade(w, r, nil)
		if err != nil {
			return
		}

		client := &websocket.Client{
			Conn: conn,
			Send: make(chan []byte, 256),
		}

		svcCtx.WSHub.Register <- client

		go client.WritePump()
		go client.ReadPump(svcCtx.WSHub)
	}
}
