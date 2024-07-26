package version

import "fmt"

const (
	major = 0
	minor = 2
	patch = 1
)

func Version() string {
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}
