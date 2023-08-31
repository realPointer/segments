package webapi

import "context"

type Disk interface {
	UploadAndReturnDownloadURL(ctx context.Context, name string, data []string) (string, error)
	IsAvailable() bool
}
