package websocket

import (
	"encoding/json"
	"net/http"

	"llm-service/internal/jwt"
	"llm-service/internal/logger"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
)

const (
	authHeader = "Authorization"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all origins for now
		return true
	},
}

type ChatHandler struct {
	grpcConn    *grpc.ClientConn
	jwtProvider *jwt.Provider
}

func NewChatHandler(grpcConn *grpc.ClientConn, jwtProvider *jwt.Provider) *ChatHandler {
	return &ChatHandler{
		grpcConn:    grpcConn,
		jwtProvider: jwtProvider,
	}
}

func (h *ChatHandler) HandleChatStream(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logger.Info(ctx, "New WebSocket connection for ChatStream")

	// Get token from header
	token := r.Header.Get(authHeader)
	if token == "" {
		// Fallback to query parameter for compatibility
		if initData := r.URL.Query().Get("init_data"); initData != "" {
			token = "tma " + initData
		} else if queryToken := r.URL.Query().Get("token"); queryToken != "" {
			token = "Bearer " + queryToken
		}
	}

	if token == "" {
		logger.Error(ctx, "No access token provided")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf(ctx, "Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Send close message
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	logger.Info(ctx, "Closing WebSocket connection")
}

func (h *ChatHandler) sendError(conn *websocket.Conn, message string) {
	errorResp := map[string]string{"error": message}
	respBytes, _ := json.Marshal(errorResp)
	conn.WriteMessage(websocket.TextMessage, respBytes)
}
