package server

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
)

// setupUIRoutes 设置前端UI路由
func (s *Server) setupUIRoutes() {
	// 定义UI路由组
	uiGroup := s.router.Group("/ui")

	// 如果启用了UI认证
	if s.config.GenerateUIAuth && s.config.UIUsername != "" && s.config.UIPassword != "" {
		// 添加基本认证中间件
		uiGroup.Use(gin.BasicAuth(gin.Accounts{
			s.config.UIUsername: s.config.UIPassword,
		}))
	}

	// UI首页
	uiGroup.GET("/", s.serveUIIndex)

	// 请求列表页面
	uiGroup.GET("/requests", s.serveUIRequestsPage)

	// 请求详情页面
	uiGroup.GET("/requests/:id", s.serveUIRequestDetailPage)

	// 配置页面
	uiGroup.GET("/config", s.serveUIConfigPage)

	// 导出页面
	uiGroup.GET("/export", s.serveUIExportPage)

	// 设置页面
	uiGroup.GET("/settings", s.serveUISettingsPage)

	// 静态资源文件
	s.router.Static("/ui/assets", "./ui/assets")
	s.router.Static("/ui/css", "./ui/css")
	s.router.Static("/ui/js", "./ui/js")

	// 设置重定向
	s.router.GET("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/ui/")
	})
}

// serveUIIndex 服务UI首页
func (s *Server) serveUIIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "index.html", gin.H{
		"title":       "Fake OpenAPI Server",
		"description": "用于记录和模拟OpenAPI请求的服务器",
	})
}

// serveUIRequestsPage 服务请求列表页面
func (s *Server) serveUIRequestsPage(c *gin.Context) {
	// 获取分页参数
	limit := 20 // 默认每页20条
	offset := 0 // 默认从0开始

	limitParam := c.Query("limit")
	offsetParam := c.Query("offset")

	if limitParam != "" {
		if _, err := fmt.Sscanf(limitParam, "%d", &limit); err != nil {
			limit = 20
		}
	}
	if offsetParam != "" {
		if _, err := fmt.Sscanf(offsetParam, "%d", &offset); err != nil {
			offset = 0
		}
	}

	// 获取请求列表
	requests, err := s.storage.GetAllRequests(limit, offset)
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "获取请求列表失败: " + err.Error(),
		})
		return
	}

	// 转换为前端模型
	var uiRequests []map[string]interface{}
	for _, req := range requests {
		uiReq := map[string]interface{}{
			"id":        req.ID,
			"method":    req.Method,
			"path":      req.Path,
			"timestamp": req.Timestamp,
		}
		uiRequests = append(uiRequests, uiReq)
	}

	c.HTML(http.StatusOK, "requests.html", gin.H{
		"title":    "请求列表",
		"requests": uiRequests,
		"limit":    limit,
		"offset":   offset,
		"hasMore":  len(requests) == limit,
	})
}

// serveUIRequestDetailPage 服务请求详情页面
func (s *Server) serveUIRequestDetailPage(c *gin.Context) {
	id := c.Param("id")
	request, err := s.storage.GetRequestByID(id)
	if err != nil {
		c.HTML(http.StatusNotFound, "error.html", gin.H{
			"error": "请求不存在: " + err.Error(),
		})
		return
	}

	// 转换为server.Request以便显示
	serverReq := convertStorageToServerRequest(request)

	c.HTML(http.StatusOK, "request_detail.html", gin.H{
		"title":   "请求详情",
		"request": serverReq,
	})
}

// serveUIConfigPage 服务配置页面
func (s *Server) serveUIConfigPage(c *gin.Context) {
	c.HTML(http.StatusOK, "config.html", gin.H{
		"title":  "服务配置",
		"config": s.config,
	})
}

// serveUIExportPage 服务导出页面
func (s *Server) serveUIExportPage(c *gin.Context) {
	// 获取存储统计信息
	stats, err := s.storage.GetStats()
	if err != nil {
		c.HTML(http.StatusInternalServerError, "error.html", gin.H{
			"error": "获取存储统计失败: " + err.Error(),
		})
		return
	}

	c.HTML(http.StatusOK, "export.html", gin.H{
		"title": "导出请求数据",
		"stats": stats,
	})
}

// serveUISettingsPage 服务设置页面
func (s *Server) serveUISettingsPage(c *gin.Context) {
	c.HTML(http.StatusOK, "settings.html", gin.H{
		"title": "服务设置",
	})
}

// LoadHTMLGlob 加载HTML模板
func (s *Server) LoadHTMLGlob(pattern string) {
	s.router.LoadHTMLGlob(pattern)
}

// LoadHTMLFiles 加载HTML文件
func (s *Server) LoadHTMLFiles(files ...string) {
	s.router.LoadHTMLFiles(files...)
}
