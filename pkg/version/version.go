package version

// Version 版本信息
const (
	Version   = "1.0.0"
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
