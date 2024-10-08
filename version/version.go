package version

import "fmt"

const (
	major = 0
	minor = 6
	patch = 5
)

func Version() string {
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}
