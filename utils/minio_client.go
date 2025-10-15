package utils

import (
	"context"
	"log"
	"os"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var MinioClient *minio.Client

func InitMinio() {
	endpoint := os.Getenv("MINIO_API")
	accessKey := os.Getenv("MINIO_KEY")
	secretKey := os.Getenv("MINIO_SECRET")
	bucketName := os.Getenv("MINIO_BUCKET")
	useSSL := false

	//log.Printf("url: %s\nKey: %s\nSecret: %s\nBucket: %s", endpoint, accessKey, secretKey, bucketName)
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
	})

	if err != nil {
		log.Fatalln("❌ Cannot connect to MinIO:", err)
	}

	// สร้าง bucket ถ้ายังไม่มี
	ctx := context.Background()
	exists, err := minioClient.BucketExists(ctx, bucketName)
	if err == nil && !exists {
		minioClient.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		log.Println("✅ Bucket created:", bucketName)
	}

	MinioClient = minioClient
	log.Println("✅ MinIO connected:", endpoint)
}
