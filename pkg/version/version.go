package version

// 版本信息变量 (可通过 -ldflags 注入)
var (
	Version   = "dev"
	BuildDate = ""
	GitCommit = ""
)

// Info 返回版本信息字符串
func Info() string {
	info := "xbuilder v" + Version
	if GitCommit != "" {
		info += " (" + GitCommit + ")"
	}
	if BuildDate != "" {
		info += " built on " + BuildDate
	}
	return info
}
