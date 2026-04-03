package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"digital-twin/src/backend/api"
	"digital-twin/src/harness/manager"
	"digital-twin/src/plugins/memory"
)

func main() {
	// 1. 加载 .env 环境变量
	if err := loadEnv(".env"); err != nil {
		log.Printf("[WARN] 加载 .env 文件失败: %v（将使用系统环境变量）", err)
	}

	// 2. 创建 HarnessManager
	configPath := "configs/harness.yaml"
	mgr, err := manager.NewHarnessManager(configPath)
	if err != nil {
		log.Fatalf("[FATAL] 创建 HarnessManager 失败: %v", err)
	}

	// 3. 启动 HarnessManager（加载配置、初始化插件、构建管道）
	if err := mgr.Start(); err != nil {
		log.Fatalf("[FATAL] 启动 HarnessManager 失败: %v", err)
	}
	log.Println("[INFO] HarnessManager 启动成功")

	// 3.5 启动记忆摘要合并定时任务
	var memorySummarizer *memory.MemorySummarizer
	if memPlugin, err := mgr.GetPlugin("memory-management"); err == nil {
		if mp, ok := memPlugin.(*memory.MemoryPlugin); ok {
			config := mgr.GetConfig()
			var llmBaseURL, llmAPIKey, llmModel string
			if config != nil {
				if memCfg, ok := config.Plugins["memory-management"]; ok {
					if v, ok := memCfg.Config["llm_base_url"]; ok {
						llmBaseURL, _ = v.(string)
					}
					if v, ok := memCfg.Config["llm_api_key"]; ok {
						llmAPIKey, _ = v.(string)
					}
					if v, ok := memCfg.Config["llm_model"]; ok {
						llmModel, _ = v.(string)
					}
				}
			}
			memorySummarizer = memory.NewMemorySummarizer(mp.GetRepo(), mp.GetStore(), mgr.GetDB(), llmBaseURL, llmAPIKey, llmModel)
			memorySummarizer.StartMemorySummarizeScheduler()
			log.Println("[INFO] 记忆摘要合并定时任务已启动")
		}
	} else if memPlugin, err := mgr.GetPlugin("memory"); err == nil {
		// 回退到类型别名
		if mp, ok := memPlugin.(*memory.MemoryPlugin); ok {
			memorySummarizer = memory.NewMemorySummarizer(mp.GetRepo(), mp.GetStore(), mgr.GetDB(), "", "", "")
			memorySummarizer.StartMemorySummarizeScheduler()
			log.Println("[INFO] 记忆摘要合并定时任务已启动")
		}
	}

	// 4. 设置路由
	router := api.SetupRouter(mgr)

	// 5. 获取服务端口
	port := os.Getenv("SERVER_PORT")
	if port == "" {
		port = "8080"
	}

	// 6. 启动 HTTP 服务
	srv := &http.Server{
		Addr:    ":" + port,
		Handler: router,
	}

	go func() {
		log.Printf("[INFO] HTTP 服务启动在 :%s", port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] HTTP 服务启动失败: %v", err)
		}
	}()

	// 7. 优雅关闭（监听 SIGINT/SIGTERM）
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("[INFO] 收到关闭信号，正在优雅关闭...")

	// 关闭 HTTP 服务
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("[ERROR] HTTP 服务关闭失败: %v", err)
	}

	// 停止记忆摘要合并定时任务
	if memorySummarizer != nil {
		memorySummarizer.Stop()
		log.Println("[INFO] 记忆摘要合并定时任务已停止")
	}

	// 停止 HarnessManager
	if err := mgr.Stop(); err != nil {
		log.Printf("[ERROR] HarnessManager 停止失败: %v", err)
	}

	log.Println("[INFO] 服务已关闭")
}

// loadEnv 从 .env 文件加载环境变量
// 支持 export VAR=VALUE 和 VAR=VALUE 两种格式
func loadEnv(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())

		// 跳过空行和注释
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		// 去掉 export 前缀
		line = strings.TrimPrefix(line, "export ")
		line = strings.TrimSpace(line)

		// 分割 key=value
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		// 去掉引号
		value = strings.Trim(value, `"'`)

		// 只在环境变量未设置时才设置（不覆盖已有值）
		if _, exists := os.LookupEnv(key); !exists {
			os.Setenv(key, value)
		}
	}

	return scanner.Err()
}
