package main

import (
	"context"
	"net/http"
	_ "net/http/pprof" // ← Import side-effect untuk pprof
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	// ← Import side-effect untuk database driver
	_ "github.com/go-sql-driver/mysql"
	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	// ← Internal packages
	"github.com/sabiqazhar/clinic-monolith/internal/config"
	"github.com/sabiqazhar/clinic-monolith/internal/infrastructure/broker"
	"github.com/sabiqazhar/clinic-monolith/internal/infrastructure/cache"
	"github.com/sabiqazhar/clinic-monolith/internal/infrastructure/db"
)

// =============================================================================
// ENTRY POINT: Aplikasi dimulai dari sini
// =============================================================================
// Blueprint Compliance:
// "Entry point bertanggung jawab atas: load config, init infrastruktur,
// dependency injection, mount routes, dan graceful shutdown."

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	// Logger (Zap)
	// Production mode: JSON logs, structured fields, async writing.
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}
	defer logger.Sync() // Flush buffer sebelum exit

	// Load Konfigurasi
	// Semua config di-externalize via environment variables.
	cfg := config.Load()

	// pprof untuk Observability (Port Terpisah)
	// Blueprint: "Endpoint observabilitas diisolasi dari API utama."
	go func() {
		logger.Info("pprof listening on :6060")
		if err := http.ListenAndServe(":6060", nil); err != nil {
			logger.Error("pprof failed", zap.Error(err))
		}
	}()

	// Inisialisasi Infrastruktur

	// PostgreSQL → Patient & Billing Modules
	// Menggunakan pgxpool untuk connection pooling yang efisien.
	coreDB, err := db.NewPostgresPool(db.PGDsn(cfg.PostgresURL))
	if err != nil {
		logger.Fatal("postgres init failed", zap.Error(err))
	}
	defer coreDB.Close() // Cleanup saat shutdown

	// MySQL → Appointment Module
	// Menggunakan db.NewMySQLDB (sudah termasuk pool config & ping)
	mysqlURL := cfg.MysqlURL
	apptDB, err := db.NewMySQLDB(db.MySQLDsn(mysqlURL))
	if err != nil {
		logger.Fatal("mysql init failed", zap.Error(err))
	}
	defer apptDB.Close()

	// RabbitMQ → Async Messaging (Transactional Outbox)
	rabbitMQ, err := broker.NewRabbitMQ(broker.RabbitURL(cfg.RabbitMQURL), logger)
	if err != nil {
		logger.Fatal("rabbitmq init failed", zap.Error(err))
	}
	defer rabbitMQ.Stop()

	// Compile-Time Dependency Injection (Wire)
	// Wire akan merakit: Repository → Service → Handler secara otomatis.
	app, err := InitializeApp(
		db.PGDsn(cfg.PostgresURL), // PostgreSQL DSN
		// db.MySQLDsn(cfg.MysqlURL),        // TODO: uncomment when appointment module ready
		cache.RedisAddr(config.BuildRedisAddr()), // Redis address
		broker.RabbitURL(cfg.RabbitMQURL),        // RabbitMQ URL
		logger,                                   // Shared logger
	)
	if err != nil {
		logger.Fatal("wire initialization failed", zap.Error(err))
	}

	// ── 6️⃣ Background Workers (Async Processing) ──
	// Context untuk graceful shutdown semua goroutine.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Outbox Relay per Database: poll → publish ke RabbitMQ.
	// Setiap DB punya relay sendiri agar isolasi failure.
	// Catatan: coreDB (PostgreSQL via pgxpool) tidak bisa langsung dipakai
	// oleh OutboxRelay yang butuh *sql.DB. Relay MySQL saja yang aktif.
	go broker.NewOutboxRelay(apptDB, rabbitMQ, logger).Start(ctx)

	// Consumer/Subscriber bisa di-start di sini jika sudah diimplementasi:
	// if app.NotificationConsumer != nil {
	// 	go app.NotificationConsumer.Start(ctx)
	// }

	// ── 7️⃣ HTTP Server (Gin Framework) ──
	gin.SetMode(gin.ReleaseMode) // Disable debug logs di production
	r := gin.Default()

	// Optional: Tambahkan middleware global di sini
	// r.Use(middleware.RequestID(), middleware.Logger(logger), middleware.Recovery())

	// Versioning & base path untuk API
	v1 := r.Group("/api/v1")

	// Mount routes per modul (Handler Self-Registration Pattern)
	// Blueprint: "Handler tahu route-nya sendiri, main.go cuma mount."
	app.PatientHandler.RegisterRoutes(v1.Group("/patients"))
	// app.AppointmentHandler.RegisterRoutes(v1.Group("/appointments"))
	// app.BillingHandler.RegisterRoutes(v1.Group("/billing"))

	// Health check endpoint (bisa untuk Kubernetes liveness probe)
	r.GET("/healthz", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	// Konfigurasi HTTP server
	serverAddr := cfg.ServerPort
	if serverAddr == "" {
		serverAddr = ":8080"
	} else if !strings.HasPrefix(serverAddr, ":") {
		serverAddr = ":" + serverAddr
	}
	srv := &http.Server{
		Addr:    serverAddr,
		Handler: r,
		// Optional: Tambahkan Read/Write timeout untuk security
		// ReadTimeout:  15 * time.Second,
		// WriteTimeout: 15 * time.Second,
		// IdleTimeout:  60 * time.Second,
	}

	// Start server di goroutine agar tidak blocking
	go func() {
		logger.Info("http server started", zap.String("addr", serverAddr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("http server error", zap.Error(err))
		}
	}()

	// Graceful Shutdown (Blueprint Requirement)
	// Menunggu sinyal interrupt (Ctrl+C) atau terminate dari OS.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit // Blocking sampai ada sinyal
	logger.Info("shutting down server...")

	// 1. Cancel context → stop background workers & relay
	cancel()

	// 2. Tunggu request HTTP yang sedang berjalan selesai (max 30 detik)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("server forced to shutdown", zap.Error(err))
	}

	// 3. Cleanup resources (defer di atas sudah handle DB.Close(), dll)
	logger.Info("server exited cleanly")
}
