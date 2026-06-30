package version

import "fmt"

var (
	Version = "v0.0.0-dev"
	Commit  = "unknown"
	Date    = "unknown"
)

func String() string {
	if Commit == "unknown" && Date == "unknown" {
		return Version
	}
	return fmt.Sprintf("%s (%s, %s)", Version, Commit, Date)
}
