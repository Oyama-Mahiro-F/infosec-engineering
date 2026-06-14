// Package main 乒乓球拍防伪溯源系统 - API服务入口
// 基于Gin框架的RESTful API服务，作为前端与区块链之间的业务逻辑桥梁
//
// 系统架构：
//   前端(小程序/Web) → API Gateway(Gin) → Blockchain Service(XuperChain SDK)
//                       ↓                    ↓
//                   PostgreSQL + Redis    7节点PBFT联盟链
//
// 启动方式：
//   go run cmd/main.go
//   docker compose up api
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/fansicheng/paddle-trace/config"
	"github.com/fansicheng/paddle-trace/internal/handler"
	"github.com/fansicheng/paddle-trace/internal/middleware"
	"github.com/fansicheng/paddle-trace/internal/service"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	log.Println("========================================")
	log.Println("  乒乓球拍防伪溯源系统 - API Service")
	log.Println("  Paddle Traceability System v1.0.0")
	log.Println("========================================")

	// ---------------------------------------------------------------------------
	// 1. 加载配置
	// ---------------------------------------------------------------------------
	cfg := config.Load()
	log.Printf("[Config] Server port: %s", cfg.Server.Port)
	log.Printf("[Config] DB: %s:%s/%s", cfg.Database.Host, cfg.Database.Port, cfg.Database.DBName)
	log.Printf("[Config] Redis: %s:%s", cfg.Redis.Host, cfg.Redis.Port)
	log.Printf("[Config] XChain nodes: %d", len(cfg.XChain.Nodes))

	// ---------------------------------------------------------------------------
	// 2. 初始化服务
	// ---------------------------------------------------------------------------

	// 2.1 区块链服务
	blockchainService, err := service.NewBlockchainService(cfg)
	if err != nil {
		log.Fatalf("[FATAL] Failed to initialize blockchain service: %v", err)
	}
	defer blockchainService.Close()

	// 2.2 NFC验证服务
	nfcService := service.NewNFCService()

	// 注册演示用NFC标签（原型阶段）
	// 生产环境中，这些密钥应在标签出厂初始化时写入并由品牌商安全管理
	log.Println("[NFC] Registering demo tags...")
	nfcService.RegisterTag("04A2B3C4D5E6F7", "00112233445566778899AABBCCDDEEFF") // demo tag 1
	nfcService.RegisterTag("04F7E6D5C4B3A2", "FFEEDDCCBBAA99887766554433221100") // demo tag 2

	// 2.3 HTTP处理器
	h := handler.NewHandler(blockchainService, nfcService)

	// ---------------------------------------------------------------------------
	// 3. 配置Gin路由
	// ---------------------------------------------------------------------------

	if cfg.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// 全局中间件
	router.Use(middleware.CORSMiddleware())

	// 健康检查（无需认证）
	router.GET("/api/v1/health", h.HealthCheck)

	// API v1 路由组
	v1 := router.Group("/api/v1")
	{
		// ---- 公开接口（消费者可访问） ----
		// 产品溯源查询
		v1.GET("/products/:id", h.GetProduct)

		// NFC真伪验证
		v1.POST("/products/verify-nfc", middleware.RateLimitMiddleware(), h.VerifyNFC)

		// ---- 认证接口 ----
		auth := v1.Group("/auth")
		{
			// TODO: 实现用户注册与登录（原型阶段可简化）
			// auth.POST("/register", h.Register)
			// auth.POST("/login", h.Login)
		}

		// ---- 需要JWT认证的接口 ----
		authRequired := v1.Group("")
		authRequired.Use(middleware.AuthMiddleware(cfg))
		{
			// 产品注册（制造商）
			authRequired.POST("/products",
				middleware.RequireRole("manufacturer", "admin"),
				h.RegisterProduct,
			)

			// 产品所有权转移
			authRequired.POST("/products/:id/transfer",
				middleware.RequireRole("manufacturer", "logistics", "distributor"),
				h.TransferProduct,
			)

			// 追加溯源记录
			authRequired.POST("/products/:id/trace",
				middleware.RequireRole("manufacturer", "logistics", "distributor"),
				h.AppendTraceRecord,
			)

			// ---- 管理员专用接口 ----
			admin := authRequired.Group("/admin")
			admin.Use(middleware.RequireRole("admin", "auditor"))
			{
				// TODO: 统计分析、异常检测等管理功能
				// admin.GET("/stats", h.GetStats)
				// admin.GET("/anomalies", h.GetAnomalies)
			}
		}
	}

	// ---------------------------------------------------------------------------
	// 4. 启动服务（支持优雅关闭）
	// ---------------------------------------------------------------------------

	srv := &http.Server{
		Addr:    ":" + cfg.Server.Port,
		Handler: router,
	}

	// 在goroutine中启动服务器
	go func() {
		log.Printf("[Server] Listening on :%s", cfg.Server.Port)
		log.Printf("[Server] API Docs: http://localhost:%s/api/v1/health", cfg.Server.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("[FATAL] Server failed: %v", err)
		}
	}()

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("[Server] Shutting down gracefully...")

	// 5秒超时优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("[FATAL] Server forced to shutdown: %v", err)
	}

	log.Println("[Server] Stopped")
}
