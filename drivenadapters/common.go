package drivenadapters

import (
	"FileEngine/common"

	"github.com/minio/minio-go/v7"
)

var (
	config      *common.Config
	minioClient *minio.Client
)

func SetConfig(c *common.Config) {
	config = c
}

func SetMinioClient(client *minio.Client) {
	minioClient = client
}
