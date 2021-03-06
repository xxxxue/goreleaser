// Package blob provides the pipe implementation that uploads files to "blob" providers, such as s3, gcs and azure.
package blob

import (
	"fmt"
	"strings"

	"github.com/goreleaser/goreleaser/internal/pipe"
	"github.com/goreleaser/goreleaser/internal/semerrgroup"
	"github.com/goreleaser/goreleaser/pkg/context"
)

// Pipe for blobs.
type Pipe struct{}

// String returns the description of the pipe.
func (Pipe) String() string {
	return "blobs"
}

// Default sets the pipe defaults.
func (Pipe) Default(ctx *context.Context) error {
	for i := range ctx.Config.Blobs {
		blob := &ctx.Config.Blobs[i]

		if blob.Bucket == "" || blob.Provider == "" {
			return fmt.Errorf("bucket 或 provider不能为空")
		}
		if blob.Folder == "" {
			blob.Folder = "{{ .ProjectName }}/{{ .Tag }}"
		}
	}
	return nil
}

// Publish to specified blob bucket url.
func (Pipe) Publish(ctx *context.Context) error {
	if len(ctx.Config.Blobs) == 0 {
		return pipe.Skip("未配置Blobs部分")
	}

	var g = semerrgroup.New(ctx.Parallelism)
	for _, conf := range ctx.Config.Blobs {
		conf := conf
		g.Go(func() error {
			return doUpload(ctx, conf)
		})
	}
	return g.Wait()
}

// errorContains check if error contains specific string.
func errorContains(err error, subs ...string) bool {
	for _, sub := range subs {
		if strings.Contains(err.Error(), sub) {
			return true
		}
	}
	return false
}
