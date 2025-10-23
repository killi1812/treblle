package minio

import (
	"context"
	"treblle/app"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"go.uber.org/zap"
)

// New creates a new minio.Client
//
// TODO: check if it should be a singleton
func New() *minio.Client {
	minioClient, err := minio.New(app.MIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(app.MIOAccessKeyID, app.MIOSecretAccessKey, ""),
		Secure: app.MIOUseSSL,
	})
	if err != nil {
		zap.S().Panicf("failed to create MinIO client %w", err)
	}

	zap.S().Info("Pinging MinIO instance")
	buckets, err := minioClient.ListBuckets(context.Background())
	if err != nil {
		zap.S().DPanicf("Error pinging minio service, err: %w", err)
		return nil
	}
	zap.S().Debugf("MinIO contains %d buckets", len(buckets))

	return minioClient
}
