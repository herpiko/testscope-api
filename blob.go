package main

import (
	"context"
	"errors"
	"io"
	"log"
	"time"

	"github.com/minio/minio-go/v7"
	uuid "github.com/satori/go.uuid"
)

type Minio struct {
	bucketName     string
	bucketLocation string
	client         *minio.Client
}

type BlobData struct {
	ID          string
	Filename    string
	ContentType string
	Size        int64
	Metadata    string
	Bucket      string
	IsPublic    bool
	Timestamp   time.Time
	HasContent  bool
}

type BlobChunk struct {
	Size int32
	Data byte
	Seq  int32
}

type BucketRequest struct {
	Filename string
	Bucket   string
}

type ID struct {
	ID string
}

const (
	CHUNK_SIZE     = 1048576 // 1 meg of buffer
	DEFAULT_BUCKET = "default-default"
	PUBLIC_BUCKET  = "public-default"
)

func (app *App) GetBlob(ctx context.Context, req *BlobData) error {
	obj, err := app.Storage.client.GetObject(ctx, req.Bucket, req.ID, minio.GetObjectOptions{})
	if err != nil {
		log.Println(err)
		return err
	}

	defer obj.Close()
	return nil
}

func (app *App) PutBlob(ctx context.Context, req *BlobData, data io.Reader) (*ID, error) {
	if req.Size == 0 {
		return nil, errors.New("empty-file")
	}
	if len(req.Bucket) == 0 {
		req.Bucket = DEFAULT_BUCKET
	}
	id := uuid.NewV4().String()

	_ = app.Storage.client.MakeBucket(ctx, req.Bucket, minio.MakeBucketOptions{
		Region: app.Storage.bucketLocation,
	})

	existing, err := app.GetFromBucket(ctx, &BucketRequest{Filename: req.Filename, Bucket: req.Bucket})
	if existing != nil {
		tag, err := app.DB.Exec(`
			UPDATE blobs
			SET timestamp=CURRENT_TIMESTAMP, size=$1, metadata=$2, content_type=$3, bucket=$4
			WHERE id=$5
			`, req.Size, req.Metadata, req.ContentType, req.Bucket, existing.ID)
		if err != nil {
			return nil, err
		}
		if _, err = tag.RowsAffected(); err != nil {
			return nil, errors.New("unable-to-update-file")
		}
		id = existing.ID
	} else {
		_, err = app.DB.Exec(`
							INSERT INTO blob_repository
							(id, filename, content_type, timestamp, size, metadata, bucket)
							VALUES
							($1, $2, $3, CURRENT_TIMESTAMP, $4, $5, $6)
						`, id, req.Filename, req.ContentType, req.Size, req.Metadata, req.Bucket)
		if err != nil {
			return nil, err
		}
	}

	tag, err := app.DB.Exec(`
		INSERT INTO blob_repository
		(id, filename, content_type, timestamp, size, metadata, bucket)
		VALUES
		($1, $2, $3, CURRENT_TIMESTAMP, $4, $5, $6)
	`, id, req.Filename, req.ContentType, req.Size, req.Metadata, req.Bucket)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	if _, err = tag.RowsAffected(); err != nil {
		return nil, errors.New("unable-to-put-file")
	}

	// Put object
	_, err = app.Storage.client.PutObject(ctx, req.Bucket, id, data, -1,
		minio.PutObjectOptions{
			ContentType: req.ContentType,
			PartSize:    10 * 1024 * 1024,
		})
	if err != nil {
		log.Println(err)
		return nil, err
	}

	return &ID{
		ID: id,
	}, nil
}

func (app *App) GetFromBucket(ctx context.Context, req *BucketRequest) (*BlobData, error) {
	rows, err := app.DB.Query(`
	
		SELECT id, filename, content_type, timestamp, size, metadata, chunks
		FROM  blob_repository
		WHERE filename like '%` + req.Filename + `' AND bucket='` + req.Bucket + `'
	`)

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	if rows.Next() {
		var resp BlobData
		var fileName string
		err = rows.Scan(&resp.ID, &resp.Filename, &resp.ContentType, &resp.Timestamp, &resp.Size, &resp.Metadata, &fileName)
		if err != nil {
			return nil, err
		}
		if fileName == resp.Filename {
			resp.HasContent = true
		}

		if req.Bucket == "public-default" {
			resp.IsPublic = true
		} else {
			resp.IsPublic = false
		}
		resp.Bucket = req.Bucket

		return &resp, nil
	}
	return nil, errors.New("blob-not-found")
}
