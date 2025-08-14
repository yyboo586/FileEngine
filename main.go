package main

import (
	"FileEngine/common"
	"FileEngine/dbaccess"
	"FileEngine/drivenadapters"
	"FileEngine/driveradapters"
	"FileEngine/interfaces"
	"FileEngine/logics"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type Server struct {
	config      *common.Config
	fileHandler interfaces.RESTHandler
}

func (s *Server) Start() {
	gin.SetMode(gin.DebugMode)

	go func() {
		server := gin.New()
		server.Use(gin.Recovery())
		server.Use(gin.Logger())

		s.fileHandler.RegisterPublic(server)

		if err := server.Run(s.config.Server.PublicAddr); err != nil {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()
}

func main() {
	config := common.NewConfig()

	log.Printf("config: %+v", config.Server)

	minioClient, err := minio.New(config.Minio.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.Minio.AccessKey, config.Minio.SecretKey, ""),
		Secure: false,
	})
	if err != nil {
		log.Fatalf("Failed to initialize Minio client: %v", err)
	}

	dbPool, err := common.NewDB(config)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	// 依赖注入
	dbaccess.SetDBPool(dbPool)

	drivenadapters.SetConfig(config)
	drivenadapters.SetMinioClient(minioClient)

	logics.SetConfig(config)

	// 控制反转
	dbFile := dbaccess.NewDBFile()

	storageAdapter := drivenadapters.NewMinioAdapter()

	logics.SetDBFile(dbFile)
	logics.SetStorageAdapter(storageAdapter)

	server := &Server{
		config:      config,
		fileHandler: driveradapters.NewFileHandler(),
	}
	server.Start()

	select {}
}
