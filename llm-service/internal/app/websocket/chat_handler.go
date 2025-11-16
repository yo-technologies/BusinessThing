package websocket

import (
	"encoding/json"
	"io"
	"net/http"

	"llm-service/internal/jwt"
	"llm-service/internal/logger"
	desc "llm-service/pkg/agent"

	"github.com/gorilla/websocket"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/protojson"
)

const (
	authHeader = "Authorization"
)

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

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
	logger.Infof(ctx, "New WebSocket connection request: method=%s, url=%s, origin=%s",
		r.Method, r.URL.String(), r.Header.Get("Origin"))

	// Get token from header
	token := r.Header.Get(authHeader)
	logger.Infof(ctx, "websocket: token from header: %t", token != "")

	if token == "" {
		// Fallback to query parameter for compatibility
		if initData := r.URL.Query().Get("init_data"); initData != "" {
			token = "tma " + initData
			logger.Info(ctx, "websocket: using init_data from query")
		} else if queryToken := r.URL.Query().Get("token"); queryToken != "" {
			// Проверяем, нет ли уже префикса в токене
			if len(queryToken) > 7 && (queryToken[:7] == "Bearer " || queryToken[:4] == "tma ") {
				token = queryToken
				logger.Infof(ctx, "websocket: using token from query (already has prefix), prefix: %s", queryToken[:min(10, len(queryToken))])
			} else {
				token = "Bearer " + queryToken
				logger.Infof(ctx, "websocket: using token from query (added Bearer prefix), prefix: %s", queryToken[:min(10, len(queryToken))])
			}
		}
	}

	if token == "" {
		logger.Error(ctx, "websocket: no access token provided (checked header and query params)")
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	logger.Infof(ctx, "websocket: attempting to upgrade connection with token prefix: %s", token[:min(20, len(token))])

	// Upgrade to WebSocket
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf(ctx, "Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	logger.Info(ctx, "websocket: connection upgraded successfully")

	// Create gRPC client with authorization metadata
	md := metadata.Pairs("authorization", token)
	grpcCtx := metadata.NewOutgoingContext(ctx, md)

	logger.Info(ctx, "websocket: creating gRPC stream client")

	// Create gRPC stream
	client := desc.NewAgentServiceClient(h.grpcConn)
	stream, err := client.StreamMessage(grpcCtx)
	if err != nil {
		logger.Errorf(ctx, "Failed to create gRPC stream: %v", err)
		h.sendError(conn, "Failed to establish connection")
		return
	}
	defer stream.CloseSend()

	logger.Info(ctx, "websocket: gRPC stream created successfully, starting message relay")

	// Channel for coordinating goroutines
	done := make(chan struct{})
	errCh := make(chan error, 2)

	// Goroutine для чтения из WebSocket и отправки в gRPC
	go func() {
		defer close(done)
		for {
			_, msgBytes, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
					logger.Errorf(ctx, "WebSocket read error: %v", err)
					errCh <- err
				}
				return
			}

			logger.Infof(ctx, "WebSocket received message: %s", string(msgBytes))

			// Десериализуем JSON в proto message используя protojson
			var req desc.StreamMessageRequest
			if err := protojson.Unmarshal(msgBytes, &req); err != nil {
				logger.Errorf(ctx, "Failed to unmarshal request: %v, raw message: %s", err, string(msgBytes))
				h.sendError(conn, "Invalid message format")
				continue
			}

			logger.Infof(ctx, "Parsed proto request: %+v", &req)

			// Отправляем в gRPC stream
			if err := stream.Send(&req); err != nil {
				logger.Errorf(ctx, "Failed to send to gRPC stream: %v", err)
				errCh <- err
				return
			}
			logger.Info(ctx, "Successfully sent request to gRPC stream")
		}
	}()

	// Goroutine для чтения из gRPC и отправки в WebSocket
	go func() {
		for {
			resp, err := stream.Recv()
			if err == io.EOF {
				logger.Info(ctx, "gRPC stream closed by server (EOF)")
				return
			}
			if err != nil {
				logger.Errorf(ctx, "Failed to receive from gRPC stream: %v", err)
				errCh <- err
				return
			}

			logger.Infof(ctx, "Received from gRPC stream: %+v", resp)

			// Сериализуем proto message в JSON используя protojson
			respBytes, err := protojson.Marshal(resp)
			if err != nil {
				logger.Errorf(ctx, "Failed to marshal response: %v", err)
				continue
			}

			logger.Infof(ctx, "Sending to WebSocket: %s", string(respBytes))

			// Отправляем в WebSocket
			if err := conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
				logger.Errorf(ctx, "Failed to write to WebSocket: %v", err)
				errCh <- err
				return
			}
		}
	}()

	// Ждем завершения или ошибки
	select {
	case <-done:
		logger.Info(ctx, "WebSocket connection closed by client")
	case err := <-errCh:
		logger.Errorf(ctx, "Connection error: %v", err)
	case <-ctx.Done():
		logger.Info(ctx, "Context cancelled")
	}

	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	logger.Info(ctx, "Closing WebSocket connection")
}

func (h *ChatHandler) sendError(conn *websocket.Conn, message string) {
	errorResp := map[string]string{"error": message}
	respBytes, _ := json.Marshal(errorResp)
	conn.WriteMessage(websocket.TextMessage, respBytes)
}
