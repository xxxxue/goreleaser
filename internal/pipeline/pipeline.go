// Package pipeline provides generic erros for pipes to use.
package pipeline

import (
	"fmt"
	"github.com/goreleaser/goreleaser/internal/pipe/before"
	"github.com/goreleaser/goreleaser/internal/pipe/defaults"
	"github.com/goreleaser/goreleaser/internal/pipe/dist"
	"github.com/goreleaser/goreleaser/internal/pipe/env"

	"github.com/goreleaser/goreleaser/internal/pipe/sourcearchive"

	"github.com/goreleaser/goreleaser/internal/pipe/archive"
	"github.com/goreleaser/goreleaser/internal/pipe/build"
	"github.com/goreleaser/goreleaser/internal/pipe/checksums"
	"github.com/goreleaser/goreleaser/internal/pipe/effectiveconfig"
	"github.com/goreleaser/goreleaser/internal/pipe/nfpm"
	"github.com/goreleaser/goreleaser/internal/pipe/sign"
	"github.com/goreleaser/goreleaser/internal/pipe/snapcraft"
	"github.com/goreleaser/goreleaser/pkg/context"
)

// Piper defines a pipe, which can be part of a pipeline (a serie of pipes).
type Piper interface {
	fmt.Stringer

	// Run the pipe
	Run(ctx *context.Context) error
}

// BuildPipeline contains all build-related pipe implementations in order.
// nolint:gochecknoglobals
var BuildPipeline = []Piper{
	env.Pipe{},             // 加载并验证环境变量
	//git.Pipe{},             // 获取并验证git repo状态
	//semver.Pipe{},          // 将当前标签解析为一个semver
	before.Pipe{},          // 在构建之前运行全局挂钩
	defaults.Pipe{},        // 加载默认配置
	//snapshot.Pipe{},        // 快照版本处理
	dist.Pipe{},            // 确保.dist是干净的
	effectiveconfig.Pipe{}, // 将实际配置（默认设置等）写入dist
	//changelog.Pipe{},       //构建发布变更日志
	build.Pipe{},           //建立
}

// Pipeline contains all pipe implementations in order.
// nolint: gochecknoglobals
var Pipeline = append(
	BuildPipeline,
	archive.Pipe{},       // 以tar.gz，zip或二进制文件存档（根本不存档）
	sourcearchive.Pipe{}, //使用git-archive归档源代码
	nfpm.Pipe{},          // 通过fpm（deb，rpm）使用“本地” go impl存档
	snapcraft.Pipe{},     // 通过snapcraft存档（快照）
	checksums.Pipe{},     // 文件的校验和
	sign.Pipe{},          // 迹象文物
	//docker.Pipe{},        // 创建并推送Docker映像
	//publish.Pipe{},       // 发表文物
)
