// Package dist provides checks to make sure the dist folder is always
// empty.
package dist

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/apex/log"
	"github.com/goreleaser/goreleaser/pkg/context"
)

// Pipe for dist.
type Pipe struct{}

func (Pipe) String() string {
	return "checking ./dist"
}

// Run the pipe.
func (Pipe) Run(ctx *context.Context) (err error) {
	_, err = os.Stat(ctx.Config.Dist)
	if os.IsNotExist(err) {
		println("./dist 不存在, 创建空文件夹:",ctx.Config.Dist)
		return mkdir(ctx)
	}
	if ctx.RmDist {
		log.Info("--rm-dist 已经设置, 开始清理dist目录")
		err = os.RemoveAll(ctx.Config.Dist)
		if err == nil {
			err = mkdir(ctx)
		}
		return err
	}
	files, err := ioutil.ReadDir(ctx.Config.Dist)
	if err != nil {
		return
	}
	if len(files) != 0 {
		log.Debugf("有 %d 个文件  on ./dist", len(files))
		return fmt.Errorf(
			"%s 不为空，请在运行goreleaser之前将其删除或使用--rm-dist标志",
			ctx.Config.Dist,
		)
	}
	log.Debug("./dist 是空的")
	return mkdir(ctx)
}

func mkdir(ctx *context.Context) error {
	// #nosec
	return os.MkdirAll(ctx.Config.Dist, 0755)
}
