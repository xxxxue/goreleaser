package cmd

import (
	"runtime"
	"time"

	"github.com/apex/log"
	"github.com/caarlos0/ctrlc"
	"github.com/fatih/color"
	"github.com/goreleaser/goreleaser/internal/middleware"
	"github.com/goreleaser/goreleaser/internal/pipeline"
	"github.com/goreleaser/goreleaser/pkg/context"
	"github.com/spf13/cobra"
)

type releaseCmd struct {
	cmd  *cobra.Command
	opts releaseOpts
}

type releaseOpts struct {
	config        string
	releaseNotes  string
	releaseHeader string
	releaseFooter string
	snapshot      bool
	skipPublish   bool
	skipSign      bool
	skipValidate  bool
	rmDist        bool
	deprecated    bool
	parallelism   int
	timeout       time.Duration
}

func newReleaseCmd() *releaseCmd {
	var root = &releaseCmd{}
	// nolint: dupl
	var cmd = &cobra.Command{
		Use:           "release",
		Aliases:       []string{"r"},
		Short:         "Releases the current project",
		SilenceUsage:  true,
		SilenceErrors: true,
		Args:          cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			start := time.Now()

			log.Infof(color.New(color.Bold).Sprint("发布中..."))

			ctx, err := releaseProject(root.opts)
			if err != nil {
				return wrapError(err, color.New(color.Bold).Sprintf("发布失败: %0.2fs", time.Since(start).Seconds()))
			}

			if ctx.Deprecated {
				log.Warn(color.New(color.Bold).Sprintf("您的配置正在使用不赞成使用的属性，请查看上面的日志以了解详细信息"))
			}

			log.Infof(color.New(color.Bold).Sprintf("发布成功 %0.2fs", time.Since(start).Seconds()))
			return nil
		},
	}

	cmd.Flags().StringVarP(&root.opts.config, "config", "f", "", "Load configuration from file 从文件加载配置")
	cmd.Flags().StringVar(&root.opts.releaseNotes, "release-notes", "", "Load custom release notes from a markdown file 从Markdown文件加载自定义发行说明")
	cmd.Flags().StringVar(&root.opts.releaseHeader, "release-header", "", "Load custom release notes header from a markdown file 从Markdown文件加载自定义发行说明标头")
	cmd.Flags().StringVar(&root.opts.releaseFooter, "release-footer", "", "Load custom release notes footer from a markdown file 从Markdown文件加载自定义发行说明页脚")
	cmd.Flags().BoolVar(&root.opts.snapshot, "snapshot", false, "Generate an unversioned snapshot release, skipping all validations and without publishing any artifacts 生成未版本化的快照版本，跳过所有验证并且不发布任何工件")
	cmd.Flags().BoolVar(&root.opts.skipPublish, "skip-publish", false, "Skips publishing artifacts 跳过发布工件")
	cmd.Flags().BoolVar(&root.opts.skipSign, "skip-sign", false, "Skips signing the artifacts 跳过对工件的签名")
	cmd.Flags().BoolVar(&root.opts.skipValidate, "skip-validate", false, "Skips several sanity checks 跳过一些健全性检查")
	cmd.Flags().BoolVar(&root.opts.rmDist, "rm-dist", false, "Remove the dist folder before building 构建之前删除dist文件夹")
	cmd.Flags().IntVarP(&root.opts.parallelism, "parallelism", "p", runtime.NumCPU(), "Amount tasks to run concurrently 并发运行的任务数量")
	cmd.Flags().DurationVar(&root.opts.timeout, "timeout", 30*time.Minute, "Timeout to the entire release process 整个发布过程超时")
	cmd.Flags().BoolVar(&root.opts.deprecated, "deprecated", false, "Force print the deprecation message - tests only 强制打印弃用消息-仅测试")
	_ = cmd.Flags().MarkHidden("deprecated")

	root.cmd = cmd
	return root
}
// 发布项目
func releaseProject(options releaseOpts) (*context.Context, error) {
	// 读取配置文件
	cfg, err := loadConfig(options.config)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.NewWithTimeout(cfg, options.timeout)
	defer cancel()

	//设置发布上下文
	setupReleaseContext(ctx, options)


	return ctx, ctrlc.Default.Run(ctx, func() error {
		for index, pipe := range pipeline.Pipeline {
			println("步骤:",index)
			if err := middleware.Logging(
				pipe.String(),
				middleware.ErrHandler(pipe.Run),
				middleware.DefaultInitialPadding,
			)(ctx); err != nil {
				return err
			}
		}
		return nil
	})
}

func setupReleaseContext(ctx *context.Context, options releaseOpts) *context.Context {
	ctx.Parallelism = options.parallelism
	log.Debugf("parallelism: %v", ctx.Parallelism)
	ctx.ReleaseNotes = options.releaseNotes
	ctx.ReleaseHeader = options.releaseHeader
	ctx.ReleaseFooter = options.releaseFooter
	ctx.Snapshot = options.snapshot
	ctx.SkipPublish = ctx.Snapshot || options.skipPublish
	ctx.SkipValidate = ctx.Snapshot || options.skipValidate
	ctx.SkipSign = options.skipSign
	ctx.RmDist = options.rmDist

	// test only
	ctx.Deprecated = options.deprecated
	return ctx
}
