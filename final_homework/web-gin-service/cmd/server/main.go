package main

import (
	"flag"
	"fmt"
	"log"

	"web-gin-service/internal/config"
	"web-gin-service/internal/grpcclient"
	"web-gin-service/internal/handler"
	"web-gin-service/internal/middleware"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
)

func main() {
	_ = godotenv.Load()
	configPath := flag.String("config", "config/config.example.yaml", "optional non-secret config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	grpcClient, err := grpcclient.New(cfg.LogicGRPC.Address)
	if err != nil {
		log.Fatalf("connect grpc: %v", err)
	}
	defer grpcClient.Close()

	h := handler.New(grpcClient)
	r := gin.Default()
	r.Use(middleware.CORS(cfg.CORS.AllowOrigins))

	api := r.Group("/api")
	{
		api.POST("/auth/register", h.Register)
		api.POST("/auth/login", h.Login)
		api.GET("/jobs", h.ListPublicJobs)
		api.GET("/jobs/:id", h.GetPublicJob)

		hr := api.Group("/hr")
		hr.Use(middleware.JWTAuth(cfg.JWT.Secret), middleware.RequireRole("hr"))
		{
			hr.GET("/jobs", h.ListHRJobs)
			hr.POST("/jobs", h.CreateJob)
			hr.PUT("/jobs/:id", h.UpdateJob)
			hr.DELETE("/jobs/:id", h.DeleteJob)
			hr.GET("/candidates", h.ListHRCandidates)
			hr.GET("/resume/download-url", h.GetDownloadURL)
			hr.POST("/ai/chat", h.Chat)
			hr.GET("/ai/history", h.GetChatHistory)
		}

		user := api.Group("/user")
		user.Use(middleware.JWTAuth(cfg.JWT.Secret), middleware.RequireRole("candidate"))
		{
			user.GET("/profile", h.GetProfile)
			user.PUT("/profile", h.UpdateProfile)
			user.GET("/resume/upload-url", h.GetUploadURL)
			user.POST("/profile/resume", h.ConfirmResume)
			user.POST("/applications", h.Apply)
		}
	}

	fmt.Printf("web-gin-service listening on %s\n", cfg.Addr())
	if err := r.Run(cfg.Addr()); err != nil {
		log.Fatalf("run: %v", err)
	}
}
