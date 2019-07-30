//nolint
package version

import (
	"fmt"
	"runtime"
)

// Variables set by build flags
var (
	Commit    = ""
	Version   = ""
	GoSumHash = ""
	BuildTags = ""
)

type versionInfo struct {
	PanaceaCore string `json:"panacea_core"`
	GitCommit   string `json:"commit"`
	GoSumHash   string `json:"go_sum_hash"`
	BuildTags   string `json:"build_tags"`
	GoVersion   string `json:"go"`
}

func (v versionInfo) String() string {
	return fmt.Sprintf(`panacea-core: %s
git commit: %s
go.sum hash: %s
build tags: %s
%s`, v.PanaceaCore, v.GitCommit, v.GoSumHash, v.BuildTags, v.GoVersion)
}

func newVersionInfo() versionInfo {
	return versionInfo{
		Version,
		Commit,
		GoSumHash,
		BuildTags,
		fmt.Sprintf("go version %s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)}
}
