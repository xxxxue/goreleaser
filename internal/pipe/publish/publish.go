// Package publish contains the publishing pipe.
package publish

import (
	"fmt"

	"github.com/goreleaser/goreleaser/internal/middleware"
	"github.com/goreleaser/goreleaser/internal/pipe/artifactory"
	"github.com/goreleaser/goreleaser/internal/pipe/blob"
	"github.com/goreleaser/goreleaser/internal/pipe/custompublishers"
	"github.com/goreleaser/goreleaser/internal/pipe/upload"
	"github.com/goreleaser/goreleaser/pkg/context"
)

// Pipe that publishes artifacts.
type Pipe struct{}

func (Pipe) String() string {
	return "publishing"
}

// Publisher should be implemented by pipes that want to publish artifacts.
type Publisher interface {
	fmt.Stringer

	// Default sets the configuration defaults
	Publish(ctx *context.Context) error
}

// nolint: gochecknoglobals
var publishers = []Publisher{
	blob.Pipe{},
	upload.Pipe{},
	custompublishers.Pipe{},
	artifactory.Pipe{},
	//docker.Pipe{},
	//docker.ManifestPipe{},
	//snapcraft.Pipe{},
	// 这应该是最后的步骤之一
	//release.Pipe{},
	// brew和scoop使用发布URL，因此，它们应该是最后一个
	//brew.Pipe{},
	//scoop.Pipe{},
	//milestone.Pipe{},
}

// Run the pipe.
func (Pipe) Run(ctx *context.Context) error {
	for _, publisher := range publishers {
		if err := middleware.Logging(
			publisher.String(),
			middleware.ErrHandler(publisher.Publish),
			middleware.ExtraPadding,
		)(ctx); err != nil {
			return fmt.Errorf("%s: 未能发布工件: %w", publisher.String(), err)
		}
	}
	return nil
}
