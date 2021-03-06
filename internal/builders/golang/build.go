package golang

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/apex/log"
	"github.com/goreleaser/goreleaser/internal/artifact"
	"github.com/goreleaser/goreleaser/internal/tmpl"
	api "github.com/goreleaser/goreleaser/pkg/build"
	"github.com/goreleaser/goreleaser/pkg/config"
	"github.com/goreleaser/goreleaser/pkg/context"
)

// Default builder instance.
// nolint: gochecknoglobals
var Default = &Builder{}

// nolint: gochecknoinits
func init() {
	api.Register("go", Default)
}

// Builder is golang builder.
type Builder struct{}

// WithDefaults sets the defaults for a golang build and returns it.
func (*Builder) WithDefaults(build config.Build) (config.Build, error) {

	// 设置编译的目录
	if build.Dir == "" {
		build.Dir = "."
	}
	if build.Main == "" {
		build.Main = "."
	}
	if len(build.Ldflags) == 0 {
		build.Ldflags = []string{"-s -w -X main.version={{.Version}} -X main.commit={{.Commit}} -X main.date={{.Date}} -X main.builtBy=goreleaser"}
	}
	if len(build.Targets) == 0 {
		if len(build.Goos) == 0 {
			build.Goos = []string{"linux", "darwin"}
		}
		if len(build.Goarch) == 0 {
			build.Goarch = []string{"amd64", "arm64", "386"}
		}
		if len(build.Goarm) == 0 {
			build.Goarm = []string{"6"}
		}
		targets, err := matrix(build)
		build.Targets = targets
		if err != nil {
			return build, err
		}
	}
	if build.GoBinary == "" {
		build.GoBinary = "go"
	}
	return build, nil
}

// Build builds a golang build.
// 构建一个  golang build
func (*Builder) Build(ctx *context.Context, build config.Build, options api.Options) error {

	// 检查是否有main方法
	if err := checkMain(build); err != nil {
		return err
	}
	target, err := newBuildTarget(options.Target)
	if err != nil {
		return err
	}

	var cmd = []string{build.GoBinary, "build"}

	var env = append(ctx.Env.Strings(), build.Env...)
	env = append(env, target.Env()...)

	artifact := &artifact.Artifact{
		Type:   artifact.Binary,
		Path:   options.Path,
		Name:   options.Name,
		Goos:   target.os,
		Goarch: target.arch,
		Goarm:  target.arm,
		Gomips: target.mips,
		Extra: map[string]interface{}{
			"Binary": strings.TrimSuffix(filepath.Base(options.Path), options.Ext),
			"Ext":    options.Ext,
			"ID":     build.ID,
		},
	}

	flags, err := processFlags(ctx, artifact, env, build.Flags, "")
	if err != nil {
		return err
	}
	cmd = append(cmd, flags...)

	asmflags, err := processFlags(ctx, artifact, env, build.Asmflags, "-asmflags=")
	if err != nil {
		return err
	}
	cmd = append(cmd, asmflags...)

	gcflags, err := processFlags(ctx, artifact, env, build.Gcflags, "-gcflags=")
	if err != nil {
		return err
	}
	cmd = append(cmd, gcflags...)

	// flag prefix is skipped because ldflags need to output a single string
	ldflags, err := processFlags(ctx, artifact, env, build.Ldflags, "")
	if err != nil {
		return err
	}
	// ldflags need to be single string in order to apply correctly
	processedLdFlags := joinLdFlags(ldflags)

	cmd = append(cmd, processedLdFlags)

	// 拼接 输出地址/名称   + 包名
	cmd = append(cmd, "-o", options.Path, build.Main)

	println("执行的命令: ")
	println()
	fmt.Printf("%v",cmd)
	println()
	println("执行的命令结束")
	// 执行拼接的 命令
	if err := run(ctx, cmd, env, build.Dir); err != nil {
		return fmt.Errorf("build 失败 %s: %w", options.Target, err)
	}

	if build.ModTimestamp != "" {
		modTimestamp, err := tmpl.New(ctx).WithEnvS(env).WithArtifact(artifact, map[string]string{}).Apply(build.ModTimestamp)
		if err != nil {
			return err
		}
		modUnix, err := strconv.ParseInt(modTimestamp, 10, 64)
		if err != nil {
			return err
		}
		modTime := time.Unix(modUnix, 0)
		err = os.Chtimes(options.Path, modTime, modTime)
		if err != nil {
			return fmt.Errorf("failed to change times for %s: %w", options.Target, err)
		}
	}

	ctx.Artifacts.Add(artifact)
	return nil
}

func processFlags(ctx *context.Context, a *artifact.Artifact, env, flags []string, flagPrefix string) ([]string, error) {
	processed := make([]string, 0, len(flags))
	for _, rawFlag := range flags {
		flag, err := tmpl.New(ctx).WithEnvS(env).WithArtifact(a, map[string]string{}).Apply(rawFlag)
		if err != nil {
			return nil, err
		}
		processed = append(processed, flagPrefix+flag)
	}
	return processed, nil
}

func joinLdFlags(flags []string) string {
	ldflagString := strings.Builder{}
	ldflagString.WriteString("-ldflags=")
	ldflagString.WriteString(strings.Join(flags, " "))

	return ldflagString.String()
}

// 执行拼接的命令
func run(ctx *context.Context, command, env []string, dir string) error {
	/* #nosec */
	var cmd = exec.CommandContext(ctx, command[0], command[1:]...)
	var log = log.WithField("env", env).WithField("cmd", command)
	cmd.Env = env
	cmd.Dir = dir
	log.Debug("running")

	// 运行命令并返回其组合的标准输出和标准错误。
	if out, err := cmd.CombinedOutput(); err != nil {
		log.WithError(err).Debug("failed")
		return errors.New(string(out))
	}
	return nil
}

type buildTarget struct {
	os, arch, arm, mips string
}

func newBuildTarget(s string) (buildTarget, error) {
	var t = buildTarget{}
	parts := strings.Split(s, "_")
	if len(parts) < 2 {
		return t, fmt.Errorf("%s 不是有效的构建目标", s)
	}
	t.os = parts[0]
	t.arch = parts[1]
	if strings.HasPrefix(t.arch, "arm") && len(parts) == 3 {
		t.arm = parts[2]
	}
	if strings.HasPrefix(t.arch, "mips") && len(parts) == 3 {
		t.mips = parts[2]
	}
	return t, nil
}

func (b buildTarget) Env() []string {
	return []string{
		"GOOS=" + b.os,
		"GOARCH=" + b.arch,
		"GOARM=" + b.arm,
		"GOMIPS=" + b.mips,
		"GOMIPS64=" + b.mips,
	}
}

func checkMain(build config.Build) error {
	var main = build.Main
	if main == "" {
		main = "."
	}
	if build.Dir != "" {
		main = filepath.Join(build.Dir, main)
	}
	stat, ferr := os.Stat(main)
	if ferr != nil {
		return ferr
	}
	if stat.IsDir() {
		packs, err := parser.ParseDir(token.NewFileSet(), main, fileFilter, 0)
		if err != nil {
			return fmt.Errorf("解析目录失败: %s: %w", main, err)
		}
		for _, pack := range packs {
			for _, file := range pack.Files {
				if hasMain(file) {
					return nil
				}
			}
		}
		return fmt.Errorf("build for %s 不包含main方法", build.Binary)
	}
	//解析单个Go源文件的源代码，并返回相应的ast
	file, err := parser.ParseFile(token.NewFileSet(), main, nil, 0)
	if err != nil {
		return fmt.Errorf("解析文件失败: %s: %w", main, err)
	}
	// 检查是否有 main 方法
	if hasMain(file) {
		return nil
	}
	return fmt.Errorf("build for %s 不包含 main function", build.Binary)
}

// TODO: can be removed once we migrate from go 1.15 to 1.16.
func fileFilter(info os.FileInfo) bool {
	return !info.IsDir()
}

func hasMain(file *ast.File) bool {
	for _, decl := range file.Decls {
		fn, isFn := decl.(*ast.FuncDecl)
		if !isFn {
			continue
		}
		if fn.Name.Name == "main" && fn.Recv == nil {
			return true
		}
	}
	return false
}
