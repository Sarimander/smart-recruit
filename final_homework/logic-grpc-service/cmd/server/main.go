package main

import (
	"flag"
	"fmt"
	"log"
	"net"

	"logic-grpc-service/internal/config"
	"logic-grpc-service/internal/db"
	grpcserver "logic-grpc-service/internal/grpc"
	osspkg "logic-grpc-service/internal/pkg/oss"
	"logic-grpc-service/internal/repository"
	"logic-grpc-service/internal/service"
	recruitv1 "logic-grpc-service/proto/gen/recruit/v1"

	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

func main() {
	_ = godotenv.Load()
	configPath := flag.String("config", "config/config.example.yaml", "optional non-secret config file")
	flag.Parse()

	cfg, err := config.Load(*configPath)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	gormDB, err := db.Connect(cfg.MySQL.DSN)
	if err != nil {
		log.Fatalf("connect db: %v", err)
	}
	if err = db.AutoMigrate(gormDB); err != nil {
		log.Fatalf("migrate: %v", err)
	}

	repo := repository.New(gormDB)
	ossClient, err := osspkg.New(cfg.OSS)
	if err != nil {
		log.Fatalf("init oss: %v", err)
	}

	aiChat, err := service.NewAIChatService(repo, cfg.DashScope)
	if err != nil {
		log.Fatalf("init ai: %v", err)
	}

	svc := service.New(repo, cfg, ossClient, aiChat)
	srv := grpcserver.New(svc)

	grpcSrv := grpc.NewServer()
	recruitv1.RegisterAuthServiceServer(grpcSrv, srv)
	recruitv1.RegisterJobServiceServer(grpcSrv, srv)
	recruitv1.RegisterCandidateServiceServer(grpcSrv, srv)
	recruitv1.RegisterApplicationServiceServer(grpcSrv, srv)
	recruitv1.RegisterOSSServiceServer(grpcSrv, srv)
	recruitv1.RegisterAIServiceServer(grpcSrv, srv)
	reflection.Register(grpcSrv)

	lis, err := net.Listen("tcp", cfg.Addr())
	if err != nil {
		log.Fatalf("listen: %v", err)
	}
	fmt.Printf("logic-grpc-service listening on %s\n", cfg.Addr())
	if err = grpcSrv.Serve(lis); err != nil {
		log.Fatalf("serve: %v", err)
	}
}
