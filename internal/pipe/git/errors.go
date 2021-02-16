package git

import (
	"errors"
	"fmt"
)

// ErrDirty happens when the repo has uncommitted/unstashed changes.
type ErrDirty struct {
	status string
}

func (e ErrDirty) Error() string {
	return fmt.Sprintf("git当前处于脏状态，请检查您的管道可以更改以下文件的内容:\n%v", e.status)
}

// ErrWrongRef happens when the HEAD reference is different from the tag being built.
type ErrWrongRef struct {
	commit, tag string
}

func (e ErrWrongRef) Error() string {
	return fmt.Sprintf("git tag %v was not made against commit %v", e.tag, e.commit)
}

// ErrNoTag happens if the underlying git repository doesn't contain any tags
// but no snapshot-release was requested.
var ErrNoTag = errors.New("git不包含任何标签。添加标签或使用--snapshot")

// ErrNotRepository happens if you try to run goreleaser against a folder
// which is not a git repository.
var ErrNotRepository = errors.New("当前文件夹不是git存储库")

// ErrNoGit happens when git is not present in PATH.
var ErrNoGit = errors.New("git不存在于PATH中")
