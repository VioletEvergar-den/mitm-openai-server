package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/llm-sec/fake-openapi-server/pkg/api"
	"github.com/spf13/cobra"
)

var (
	// 请求管理特定的标志
	requestID    string
	outputFormat string
	deleteAll    bool
)

// requestsCmd 表示requests命令
var requestsCmd = &cobra.Command{
	Use:   "requests",
	Short: "管理记录的请求",
	Long: `管理API服务器记录的请求数据。
	
此命令允许列出、查看和删除已记录的请求数据。`,
	Run: func(cmd *cobra.Command, args []string) {
		cmd.Help()
	},
}

// listCmd 表示list子命令
var listCmd = &cobra.Command{
	Use:   "list",
	Short: "列出所有记录的请求",
	Long:  `列出API服务器记录的所有请求数据。`,
	Run: func(cmd *cobra.Command, args []string) {
		listRequests()
	},
}

// getCmd 表示get子命令
var getCmd = &cobra.Command{
	Use:   "get [requestID]",
	Short: "获取特定ID的请求",
	Long:  `根据ID获取API服务器记录的特定请求数据。`,
	Run: func(cmd *cobra.Command, args []string) {
		if len(args) == 0 && requestID == "" {
			fmt.Println("错误: 请提供请求ID")
			cmd.Help()
			return
		}

		// 优先使用位置参数
		if len(args) > 0 {
			requestID = args[0]
		}

		getRequest(requestID)
	},
}

// deleteCmd 表示delete子命令
var deleteCmd = &cobra.Command{
	Use:   "delete [requestID]",
	Short: "删除特定ID的请求或所有请求",
	Long:  `删除API服务器记录的特定请求数据或所有请求数据。`,
	Run: func(cmd *cobra.Command, args []string) {
		if deleteAll {
			deleteAllRequests()
			return
		}

		if len(args) == 0 && requestID == "" {
			fmt.Println("错误: 请提供请求ID或使用--all标志")
			cmd.Help()
			return
		}

		// 优先使用位置参数
		if len(args) > 0 {
			requestID = args[0]
		}

		deleteRequest(requestID)
	},
}

func init() {
	rootCmd.AddCommand(requestsCmd)

	// 添加子命令
	requestsCmd.AddCommand(listCmd)
	requestsCmd.AddCommand(getCmd)
	requestsCmd.AddCommand(deleteCmd)

	// 设置共享标志
	requestsCmd.PersistentFlags().StringVarP(&outputFormat, "output", "o", "json", "输出格式 (json, yaml, table)")

	// get命令的标志
	getCmd.Flags().StringVarP(&requestID, "id", "i", "", "请求ID")

	// delete命令的标志
	deleteCmd.Flags().StringVarP(&requestID, "id", "i", "", "请求ID")
	deleteCmd.Flags().BoolVar(&deleteAll, "all", false, "删除所有请求")
}

// 获取存储实例
func getStorage() (api.Storage, error) {
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		return nil, fmt.Errorf("无法获取数据目录的绝对路径: %w", err)
	}

	if err := os.MkdirAll(absDataDir, 0755); err != nil {
		return nil, fmt.Errorf("创建数据目录失败: %w", err)
	}

	return api.NewFileStorage(absDataDir)
}

// 打印请求数据
func printRequest(req *api.Request) {
	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(req, "", "  ")
		if err != nil {
			log.Fatalf("无法序列化请求: %v", err)
		}
		fmt.Println(string(data))
	case "yaml":
		fmt.Println("YAML输出尚未实现")
	case "table":
		fmt.Printf("ID: %s\n", req.ID)
		fmt.Printf("时间戳: %s\n", req.Timestamp)
		fmt.Printf("方法: %s\n", req.Method)
		fmt.Printf("路径: %s\n", req.Path)
		fmt.Printf("IP地址: %s\n", req.IPAddress)

		fmt.Println("\n请求头:")
		for k, v := range req.Headers {
			fmt.Printf("  %s: %s\n", k, v)
		}

		fmt.Println("\n查询参数:")
		for k, v := range req.Query {
			fmt.Printf("  %s: %s\n", k, v)
		}

		if req.Body != nil {
			fmt.Println("\n请求体:")
			bodyData, err := json.MarshalIndent(req.Body, "  ", "  ")
			if err != nil {
				fmt.Printf("  无法序列化请求体: %v\n", err)
			} else {
				fmt.Println("  " + string(bodyData))
			}
		}
	default:
		log.Fatalf("不支持的输出格式: %s", outputFormat)
	}
}

// 列出所有请求
func listRequests() {
	storage, err := getStorage()
	if err != nil {
		log.Fatalf("初始化存储失败: %v", err)
	}

	requests, err := storage.GetAllRequests()
	if err != nil {
		log.Fatalf("获取请求列表失败: %v", err)
	}

	if len(requests) == 0 {
		fmt.Println("没有找到请求记录")
		return
	}

	switch outputFormat {
	case "json":
		data, err := json.MarshalIndent(requests, "", "  ")
		if err != nil {
			log.Fatalf("无法序列化请求: %v", err)
		}
		fmt.Println(string(data))
	case "yaml":
		fmt.Println("YAML输出尚未实现")
	case "table":
		fmt.Printf("找到 %d 个请求记录:\n\n", len(requests))
		fmt.Printf("%-36s  %-24s  %-7s  %s\n", "ID", "时间戳", "方法", "路径")
		fmt.Println(strings.Repeat("-", 80))

		for _, req := range requests {
			fmt.Printf("%-36s  %-24s  %-7s  %s\n", req.ID, req.Timestamp, req.Method, req.Path)
		}
	default:
		log.Fatalf("不支持的输出格式: %s", outputFormat)
	}
}

// 获取特定请求
func getRequest(id string) {
	storage, err := getStorage()
	if err != nil {
		log.Fatalf("初始化存储失败: %v", err)
	}

	request, err := storage.GetRequestByID(id)
	if err != nil {
		log.Fatalf("获取请求失败: %v", err)
	}

	printRequest(request)
}

// 删除特定请求
func deleteRequest(id string) {
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		log.Fatalf("无法获取数据目录的绝对路径: %v", err)
	}

	filePath := filepath.Join(absDataDir, fmt.Sprintf("%s.json", id))
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		log.Fatalf("请求ID不存在: %s", id)
	}

	if err := os.Remove(filePath); err != nil {
		log.Fatalf("删除请求失败: %v", err)
	}

	fmt.Printf("成功删除请求: %s\n", id)
}

// 删除所有请求
func deleteAllRequests() {
	absDataDir, err := filepath.Abs(dataDir)
	if err != nil {
		log.Fatalf("无法获取数据目录的绝对路径: %v", err)
	}

	files, err := os.ReadDir(absDataDir)
	if err != nil {
		log.Fatalf("读取数据目录失败: %v", err)
	}

	count := 0
	for _, file := range files {
		if !file.IsDir() && strings.HasSuffix(file.Name(), ".json") {
			filePath := filepath.Join(absDataDir, file.Name())
			if err := os.Remove(filePath); err != nil {
				log.Printf("警告: 无法删除文件 %s: %v", file.Name(), err)
				continue
			}
			count++
		}
	}

	fmt.Printf("成功删除 %d 个请求记录\n", count)
}
