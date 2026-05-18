package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
	"kirimaja-go/internal/common/email"
	"kirimaja-go/internal/common/midtrans"
	"kirimaja-go/internal/common/opencage"
	"kirimaja-go/internal/common/pdf"
	"kirimaja-go/internal/common/qrcode"
	"kirimaja-go/internal/common/worker"
	"kirimaja-go/internal/config"
	"kirimaja-go/internal/database"
	"kirimaja-go/internal/middleware"
	"kirimaja-go/internal/modules/auth"
	"kirimaja-go/internal/modules/branches"
	employee_branches "kirimaja-go/internal/modules/employee_branches"
	"kirimaja-go/internal/modules/history"
	"kirimaja-go/internal/modules/permissions"
	"kirimaja-go/internal/modules/profile"
	"kirimaja-go/internal/modules/roles"
	"kirimaja-go/internal/modules/shipments"
	"kirimaja-go/internal/modules/shipments/branch"
	"kirimaja-go/internal/modules/shipments/courier"
	"kirimaja-go/internal/modules/shipments/webhook"
	user_addresses "kirimaja-go/internal/modules/user_addresses"
)

func main() {
	cfg := config.Load()
	cfg.Validate()

	db, err := database.Connect(cfg.DatabaseURL, cfg.IsProduction())
	if err != nil {
		log.Fatalf("Database connection failed: %v", err)
	}

	// Common clients
	geoCli := opencage.New(cfg.OpenCageAPIKey)
	mtCli := midtrans.New(cfg.MidtransServerKey, cfg.MidtransEnv)
	qrSvc := qrcode.New(cfg.PublicDir)

	// Email + worker
	smtpPort := 587
	fmt.Sscanf(cfg.SMTPPort, "%d", &smtpPort)
	emailSvc := email.New(cfg.SMTPHost, smtpPort, cfg.SMTPUser, cfg.SMTPPass, cfg.SMTPSender)
	workerClient := worker.NewClient(cfg.RedisURL)
	workerServer := worker.NewServer(cfg.RedisURL, emailSvc, db)
	workerServer.Start()
	defer workerServer.Shutdown()

	// PDF
	pdfSvc, closePDF := pdf.New(cfg.PublicDir)
	defer closePDF()

	// Middleware
	authMw := middleware.AuthRequired(cfg.JWTSecret)
	permMw := middleware.RequirePermission(db)

	// Auth
	authRepo := auth.NewRepository(db)
	authSvc := auth.NewService(authRepo, cfg.JWTSecret, cfg.JWTExpiresIn)
	authHdlr := auth.NewHandler(authSvc)

	// Roles
	rolesRepo := roles.NewRepository(db)
	rolesSvc := roles.NewService(rolesRepo)
	rolesHdlr := roles.NewHandler(rolesSvc)

	// Permissions
	permsRepo := permissions.NewRepository(db)
	permsSvc := permissions.NewService(permsRepo)
	permsHdlr := permissions.NewHandler(permsSvc)

	// Branches
	branchesRepo := branches.NewRepository(db)
	branchesSvc := branches.NewService(branchesRepo)
	branchesHdlr := branches.NewHandler(branchesSvc)

	// Employee Branches
	ebRepo := employee_branches.NewRepository(db)
	ebSvc := employee_branches.NewService(ebRepo)
	ebHdlr := employee_branches.NewHandler(ebSvc)

	// Profile
	profileRepo := profile.NewRepository(db)
	profileSvc := profile.NewService(profileRepo)
	profileHdlr := profile.NewHandler(profileSvc, cfg.PublicDir)

	// User Addresses
	uaRepo := user_addresses.NewRepository(db)
	uaSvc := user_addresses.NewService(uaRepo, geoCli)
	uaHdlr := user_addresses.NewHandler(uaSvc, cfg.PublicDir)

	// Shipments
	shipmentsRepo := shipments.NewRepository(db)
	shipmentsSvc := shipments.NewService(shipmentsRepo, geoCli, mtCli, qrSvc, workerClient, pdfSvc)
	shipmentsHdlr := shipments.NewHandler(shipmentsSvc)
	webhookHdlr := webhook.NewHandler(shipmentsSvc)
	branchHdlr := branch.NewHandler(shipmentsSvc)
	courierHdlr := courier.NewHandler(shipmentsSvc, cfg.PublicDir)

	r := gin.Default()
	r.Static("/uploads", cfg.PublicDir+"/uploads")

	api := r.Group("/api/v1")
	auth.RegisterRoutes(api, authHdlr)
	roles.RegisterRoutes(api, rolesHdlr, authMw, permMw)
	permissions.RegisterRoutes(api, permsHdlr, authMw, permMw)
	branches.RegisterRoutes(api, branchesHdlr, authMw, permMw)
	employee_branches.RegisterRoutes(api, ebHdlr, authMw, permMw)
	profile.RegisterRoutes(api, profileHdlr, authMw)
	user_addresses.RegisterRoutes(api, uaHdlr, authMw)
	shipments.RegisterRoutes(api, shipmentsHdlr, authMw, permMw)
	webhook.RegisterRoutes(api, webhookHdlr, middleware.RateLimit(rate.Limit(5), 10))
	branch.RegisterRoutes(api, branchHdlr, authMw, permMw)
	courier.RegisterRoutes(api, courierHdlr, authMw, permMw)

	historyRepo := history.NewRepository(db)
	historySvc := history.NewService(historyRepo)
	historyHdlr := history.NewHandler(historySvc)
	history.RegisterRoutes(api, historyHdlr, authMw, permMw)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	srv := &http.Server{Addr: ":" + cfg.Port, Handler: r}

	go func() {
		log.Printf("Server running on port %s", cfg.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Server failed: %v", err)
		}
	}()

	// Graceful shutdown: on SIGINT/SIGTERM stop accepting new requests and let
	// in-flight ones (e.g. a payment write) finish before the deferred worker
	// and Chrome cleanup run.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	<-ctx.Done()
	stop()
	log.Println("Shutting down...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()
	if err := srv.Shutdown(shutdownCtx); err != nil {
		log.Printf("graceful shutdown failed: %v", err)
	}
}
