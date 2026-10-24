// Package storage — хранилища для фото питомцев: локальный диск (по умолчанию)
// и S3-совместимое Yandex Object Storage (если заданы ключи).
package storage

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func randomName(ext string) (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b) + ext, nil
}

// Local складывает файлы в каталог и отдаёт их по пути /uploads/<имя>.
type Local struct {
	Dir string
}

func NewLocal(dir string) (*Local, error) {
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	return &Local{Dir: dir}, nil
}

func (l *Local) Save(_ context.Context, data io.Reader, ext, _ string) (string, error) {
	name, err := randomName(ext)
	if err != nil {
		return "", err
	}
	f, err := os.Create(filepath.Join(l.Dir, name))
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, data); err != nil {
		return "", err
	}
	return "/uploads/" + name, nil
}

// S3 — Yandex Object Storage.
type S3 struct {
	client *s3.Client
	bucket string
}

func NewS3(accessKey, secretKey, bucket string) *S3 {
	cfg := aws.Config{
		Region:       "ru-central1",
		Credentials:  credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		BaseEndpoint: aws.String("https://storage.yandexcloud.net"),
	}
	return &S3{client: s3.NewFromConfig(cfg), bucket: bucket}
}

func (s *S3) Save(ctx context.Context, data io.Reader, ext, contentType string) (string, error) {
	name, err := randomName(ext)
	if err != nil {
		return "", err
	}
	key := "photos/" + name
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      &s.bucket,
		Key:         &key,
		Body:        data,
		ContentType: &contentType,
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("https://storage.yandexcloud.net/%s/%s", s.bucket, key), nil
}
