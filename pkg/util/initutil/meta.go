package initutil

import "fmt"

// Build metadata. These values are intended to be overridden by values
// supplied by the build environment at link time. Values are specified
// by passing "-X package.Varname=Value" to the linker. For example:
//
//  go build -ldflags="-X meta.Name=BFF -X main.Version=1.0" ...etc...
//
// Ref: https://golang.org/cmd/link/
var (
	ApplicationName string
	Version         string
)

type BuildMetadata struct {
	ApplicationName string
	Version         string
}

func (b *BuildMetadata) String() string {
	return fmt.Sprintf(
		"app=%s version=%s",
		b.ApplicationName,
		b.Version)
}

func NewBuildMetadata() *BuildMetadata {
	return &BuildMetadata{
		ApplicationName: ApplicationName,
		Version:         Version,
	}
}
