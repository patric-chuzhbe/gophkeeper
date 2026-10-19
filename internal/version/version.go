package version

var (
	Version   = "dev"
	BuildDate = "unknown"
)

func Info() string {
	return "Version: " + Version + "\nBuild date: " + BuildDate
}
