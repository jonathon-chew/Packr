package platform

import (
	"fmt"
	"runtime"
)

func Key() string {
	return fmt.Sprintf("%s-%s", runtime.GOOS, runtime.GOARCH)
}
