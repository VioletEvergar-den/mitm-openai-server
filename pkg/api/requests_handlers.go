package api

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// GetRequests 获取请求列表
func (s *UIServer) GetRequests(c *gin.Context) {
	// 处理分页参数
	page, size := GetPaginationParams(c)
	limit := size
	offset := (page - 1) * size

	// 从存储中获取请求列表
	requests, err := s.storage.GetAllRequests(limit, offset)
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Code: 10012,
			Msg:  "获取请求列表失败: " + err.Error(),
		})
		return
	}

	// 获取总请求数
	var total int64
	total = int64(len(requests))

	// 将存储请求转换为服务器请求
	serverRequests := make([]*Request, len(requests))
	for i, req := range requests {
		serverRequests[i] = ConvertStorageToAPIRequest(req)
	}

	// 返回请求列表和分页信息
	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "获取请求列表成功",
		Data: map[string]interface{}{
			"requests": serverRequests,
			"total":    total,
			"page":     page,
			"size":     size,
		},
	})
}

// GetRequestByID 获取特定ID的请求
func (s *UIServer) GetRequestByID(c *gin.Context) {
	// 获取请求ID
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Code: 10003,
			Msg:  "请求ID不能为空",
		})
		return
	}

	// 从存储中获取请求
	req, err := s.storage.GetRequestByID(id)
	if err != nil {
		c.JSON(http.StatusNotFound, StandardResponse{
			Code: 10004,
			Msg:  "请求不存在: " + err.Error(),
		})
		return
	}

	// 转换为服务器请求模型
	serverReq := ConvertStorageToAPIRequest(req)

	// 返回请求详情
	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "请求获取成功",
		Data: serverReq,
	})
}

// DeleteRequest 删除特定ID的请求
func (s *UIServer) DeleteRequest(c *gin.Context) {
	// 获取请求ID
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, StandardResponse{
			Code: 10006,
			Msg:  "请求ID不能为空",
		})
		return
	}

	// 从存储中删除请求
	err := s.storage.DeleteRequest(id)
	if err != nil {
		statusCode := http.StatusInternalServerError

		c.JSON(statusCode, StandardResponse{
			Code: 10007,
			Msg:  "删除请求失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "请求已成功删除",
	})
}

// DeleteAllRequests 删除所有请求
func (s *UIServer) DeleteAllRequests(c *gin.Context) {
	// 从存储中删除所有请求
	err := s.storage.DeleteAllRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Code: 10008,
			Msg:  "删除所有请求失败: " + err.Error(),
		})
		return
	}

	// 返回成功响应
	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "所有请求已成功删除",
	})
}

// ExportRequests 导出请求为JSONL格式
func (s *UIServer) ExportRequests(c *gin.Context) {
	// 从存储中导出请求
	filePath, err := s.storage.ExportRequests()
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Code: 10009,
			Msg:  "导出请求失败: " + err.Error(),
		})
		return
	}

	// 返回导出文件路径
	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "请求已成功导出",
		Data: map[string]string{
			"file_path": filePath,
		},
	})
}

// GetStorageStats 获取存储统计信息
func (s *UIServer) GetStorageStats(c *gin.Context) {
	// 获取存储统计信息
	stats, err := s.storage.GetStats()
	if err != nil {
		c.JSON(http.StatusInternalServerError, StandardResponse{
			Code: 10010,
			Msg:  "获取存储统计信息失败: " + err.Error(),
		})
		return
	}

	// 返回统计信息
	c.JSON(http.StatusOK, StandardResponse{
		Code: 0,
		Msg:  "获取存储统计信息成功",
		Data: stats,
	})
}
