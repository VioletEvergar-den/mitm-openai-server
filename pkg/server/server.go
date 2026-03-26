package server

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/llm-sec/mitm-openai-server/pkg/api"
	"github.com/llm-sec/mitm-openai-server/pkg/logger"
	"github.com/llm-sec/mitm-openai-server/pkg/openai"
)

type Server struct {
	config        api.ServerConfig
	router        *gin.Engine
	uiServer      api.UIServerInterface
	openaiHandler *openai.Handler
	openaiService openai.Service
	defaultUserID int64
	storagePath   string
}

func NewServerWithConfig(config api.ServerConfig) *Server {
	gin.SetMode(gin.ReleaseMode)

	router := gin.New()
	router.Use(gin.Recovery())

	if config.EnableCORS {
		router.Use(corsMiddleware(config.AllowOrigins))
	}

	logger.InitLogBuffer(1000)

	s := &Server{
		config: config,
		router: router,
	}

	s.openaiService = openai.NewService(openai.Config{
		ProxyMode:      config.ProxyMode,
		TargetURL:      config.TargetURL,
		TargetAuthType: config.TargetAuthType,
		TargetUsername: config.TargetUsername,
		TargetPassword: config.TargetPassword,
		TargetToken:    config.TargetToken,
		ModelMapping:   config.ModelMapping,
	})

	s.openaiHandler = openai.InitGlobalHandler(config.Storage, s.openaiService)

	s.defaultUserID = s.getDefaultUserID()

	s.uiServer = api.NewUIServer(config.Storage, config, s.openaiService)

	authMiddleware := s.createAuthMiddleware()
	apiMiddleware := s.createAPIMiddleware()

	s.openaiHandler.SetupRoutes(router, apiMiddleware)

	s.uiServer.SetupUIRoutes(router, authMiddleware, apiMiddleware)

	s.setupStaticRoutes(router)

	s.setupFallbackRoutes(router, apiMiddleware)

	return s
}

func (s *Server) getDefaultUserID() int64 {
	if s.config.Storage == nil {
		return 1
	}

	user, err := s.config.Storage.GetUserByUsername("api_user")
	if err != nil {
		fmt.Printf("警告: 无法获取默认API用户: %v，使用ID=1\n", err)
		return 1
	}

	return user.ID
}

func (s *Server) createAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !s.config.EnableAuth {
			c.Next()
			return
		}

		switch s.config.AuthType {
		case "basic":
			username, password, hasAuth := c.Request.BasicAuth()
			if !hasAuth || username != s.config.Username || password != s.config.Password {
				c.Header("WWW-Authenticate", `Basic realm="API Authentication"`)
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": map[string]interface{}{
						"message": "Authentication required",
						"type":    "invalid_request_error",
						"code":    "unauthorized",
					},
				})
				c.Abort()
				return
			}
		case "token":
			authHeader := c.GetHeader("Authorization")
			if authHeader == "" {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": map[string]interface{}{
						"message": "Authorization header required",
						"type":    "invalid_request_error",
						"code":    "unauthorized",
					},
				})
				c.Abort()
				return
			}

			token := authHeader
			if strings.HasPrefix(strings.ToLower(authHeader), "bearer ") {
				token = authHeader[7:]
			}

			if token != s.config.Token {
				c.JSON(http.StatusUnauthorized, gin.H{
					"error": map[string]interface{}{
						"message": "Invalid token",
						"type":    "invalid_request_error",
						"code":    "unauthorized",
					},
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

func (s *Server) createAPIMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", s.defaultUserID)
		c.Set("username", "api_user")
		c.Next()
	}
}

func (s *Server) setupStaticRoutes(router *gin.Engine) {
	uiDir := s.config.UIDir
	if uiDir == "" {
		uiDir = "./react-ui/dist"
	}

	absUIDir, err := filepath.Abs(uiDir)
	if err != nil {
		absUIDir = uiDir
	}

	if _, err := os.Stat(absUIDir); err == nil {
		router.Static("/ui/assets", filepath.Join(absUIDir, "assets"))
		router.StaticFile("/ui/favicon.ico", filepath.Join(absUIDir, "favicon.ico"))

		router.GET("/ui", func(c *gin.Context) {
			c.File(filepath.Join(absUIDir, "index.html"))
		})

		router.GET("/ui/login", func(c *gin.Context) {
			c.File(filepath.Join(absUIDir, "index.html"))
		})

		router.GET("/ui/requests", func(c *gin.Context) {
			c.File(filepath.Join(absUIDir, "index.html"))
		})

		router.GET("/ui/settings", func(c *gin.Context) {
			c.File(filepath.Join(absUIDir, "index.html"))
		})

		router.GET("/ui/guide", func(c *gin.Context) {
			c.File(filepath.Join(absUIDir, "index.html"))
		})
	}
}

func (s *Server) setupFallbackRoutes(router *gin.Engine, apiMiddleware gin.HandlerFunc) {
	router.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/v1/") || strings.HasPrefix(path, "/v2/") {
			apiMiddleware(c)
			if !c.IsAborted() {
				s.openaiHandler.HandleRequest(c)
			}
			return
		}

		if strings.HasPrefix(path, "/ui/") {
			uiDir := s.config.UIDir
			if uiDir == "" {
				uiDir = "./react-ui/dist"
			}
			indexPath := filepath.Join(uiDir, "index.html")
			if _, err := os.Stat(indexPath); err == nil {
				c.File(indexPath)
				return
			}
		}

		c.JSON(http.StatusNotFound, gin.H{
			"error": map[string]interface{}{
				"message": fmt.Sprintf("Not found: %s", path),
				"type":    "invalid_request_error",
				"code":    "not_found",
			},
		})
	})
}

func (s *Server) Run(addr string) error {
	return s.router.Run(addr)
}

func (s *Server) GetConfig() api.ServerConfig {
	return s.config
}

func corsMiddleware(allowOrigins string) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		if origin == "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", allowOrigins)
		} else {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		}

		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, Authorization, X-Requested-With")
		c.Writer.Header().Set("Access-Control-Max-Age", "86400")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
