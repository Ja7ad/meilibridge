package version

import "fmt"

const (
	major = 0
	minor = 3
	patch = 0
)

func Version() string {
	return fmt.Sprintf("v%d.%d.%d", major, minor, patch)
}
