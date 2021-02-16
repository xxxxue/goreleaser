package cmd

import (
	"os"

	"github.com/apex/log"
	"github.com/goreleaser/goreleaser/pkg/config"
)

// 加载yaml配置文件
func loadConfig(path string) (config.Project, error) {
	if path != "" {
		return config.Load(path)
	}
	for _, f := range [4]string{
		".goreleaser.yml",
		".goreleaser.yaml",
		"goreleaser.yml",
		"goreleaser.yaml",
	} {
		proj, err := config.Load(f)
		if err != nil && os.IsNotExist(err) {
			continue
		}
		return proj, err
	}
	// the user didn't specify a config file and the known possible file names
	// don't exist, so, return an empty config and a nil err.
	//用户未指定配置文件和已知的可能的文件名
	//不存在，因此，返回空配置和nil err。
	log.Warn("使用默认值找不到配置文件...")
	return config.Project{}, nil
}
