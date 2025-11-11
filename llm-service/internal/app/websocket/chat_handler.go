package websocket

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"sync"

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

	// Validate token (optional validation, will be validated by gRPC interceptor as well)
	// The token validation is primarily handled by the gRPC stream interceptor

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		logger.Errorf(ctx, "Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	// Create gRPC client
	client := desc.NewChatServiceClient(h.grpcConn)

	// Create context with metadata
	md := metadata.New(map[string]string{"authorization": token})
	streamCtx, cancel := context.WithCancel(metadata.NewOutgoingContext(ctx, md))
	defer cancel()

	// Start bidirectional stream
	stream, err := client.ChatStream(streamCtx)
	if err != nil {
		logger.Errorf(ctx, "Failed to start stream: %v", err)
		h.sendError(conn, "Failed to start chat stream")
		return
	}

	// WaitGroup to coordinate goroutines
	var wg sync.WaitGroup
	errChan := make(chan error, 2)

	// Goroutine to read from WebSocket and send to gRPC stream
	wg.Add(1)
	go func() {
		defer wg.Done()
		defer func() {
			if err := stream.CloseSend(); err != nil {
				logger.Errorf(ctx, "Failed to close send stream: %v", err)
			}
		}()

		for {
			messageType, message, err := conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					logger.Info(ctx, "WebSocket closed normally")
					return
				}
				logger.Errorf(ctx, "Failed to read from WebSocket: %v", err)
				errChan <- err
				return
			}

			if messageType != websocket.TextMessage {
				logger.Errorf(ctx, "Invalid message type: %d", messageType)
				continue
			}

			// Parse message as ChatStreamRequest
			var req desc.ChatStreamRequest
			if err := protojson.Unmarshal(message, &req); err != nil {
				logger.Errorf(ctx, "Failed to unmarshal request: %v", err)
				h.sendError(conn, "Invalid request format")
				continue
			}

			// Send to gRPC stream
			if err := stream.Send(&req); err != nil {
				logger.Errorf(ctx, "Failed to send to gRPC stream: %v", err)
				errChan <- err
				return
			}
		}
	}()

	// Goroutine to read from gRPC stream and send to WebSocket
	wg.Add(1)
	go func() {
		defer wg.Done()

		for {
			resp, err := stream.Recv()
			if err != nil {
				if err == io.EOF {
					logger.Info(ctx, "gRPC stream ended")
					return
				}
				logger.Errorf(ctx, "Failed to receive from gRPC stream: %v", err)
				errChan <- err
				return
			}

			// Send response to WebSocket
			respBytes, err := protojson.Marshal(resp)
			if err != nil {
				logger.Errorf(ctx, "Failed to marshal response: %v", err)
				h.sendError(conn, "Internal error")
				continue
			}

			if err := conn.WriteMessage(websocket.TextMessage, respBytes); err != nil {
				logger.Errorf(ctx, "Failed to write to WebSocket: %v", err)
				errChan <- err
				return
			}
		}
	}()

	// Wait for either goroutine to finish or error
	go func() {
		wg.Wait()
		close(errChan)
	}()

	// Block until error or completion
	for err := range errChan {
		if err != nil {
			logger.Errorf(ctx, "Stream error: %v", err)
			break
		}
	}

	// Send close message
	conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
	logger.Info(ctx, "Closing WebSocket connection")
}

func (h *ChatHandler) sendError(conn *websocket.Conn, message string) {
	errorResp := map[string]string{"error": message}
	respBytes, _ := json.Marshal(errorResp)
	conn.WriteMessage(websocket.TextMessage, respBytes)
}
