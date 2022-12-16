package main

import (
	"lab_2_KBRS/handlers"

	"github.com/gin-gonic/gin"
	socketio "github.com/googollee/go-socket.io"
	"go.uber.org/zap"
)

func InitLogger() *zap.SugaredLogger {
	logger, _ := zap.NewProduction()
	sugarLogger := logger.Sugar()
	return sugarLogger
}

func InitSocketServer(server *socketio.Server, h *handlers.Handler, logger *zap.SugaredLogger) {
	server.OnConnect("/", func(s socketio.Conn) error {
		s.SetContext("")
		logger.Infof("user with ID - %v is connected", s.ID())
		return nil
	})

	server.OnEvent("/", "login", h.Login)
	server.OnEvent("/auth", "createFile", h.CreateFile)
	server.OnEvent("/auth", "getFile", h.GetFile)
	server.OnEvent("/auth", "editFile", h.UpdateFile)
	server.OnEvent("/auth", "deleteFile", h.DeleteFile)

	server.OnError("/", func(s socketio.Conn, e error) {
		logger.Errorf("error : %v", e)
	})

	server.OnDisconnect("/", func(s socketio.Conn, reason string) {
		logger.Infof("user with ID - %v closed conn due to %v", s.ID(), reason)
	})
}

func main() {
	logger := InitLogger()
	defer logger.Sync()

	router := gin.New()
	server := socketio.NewServer(nil)
	h := handlers.NewHandler(logger)
	InitSocketServer(server, h, logger)

	go func() {
		if err := server.Serve(); err != nil {
			logger.Fatalf("error occurred while running socket.io server : %v", err)
		}
	}()
	defer server.Close()

	router.GET("/socket", gin.WrapH(server))

	if err := router.Run(":8000"); err != nil {
		logger.Fatalf("error occurred while running app : %v", err)
	}
}
