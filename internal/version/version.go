package version

import "log"

var (
	buildVersion string
	buildCommit  string
)

func DumpVersion(l *log.Logger) {
	l.Printf("----------------------------------------\n")
	l.Printf("Version: %s\n", buildVersion)
	l.Printf("Build commit hash: %s\n", buildCommit)
	l.Printf("----------------------------------------\n")
}
