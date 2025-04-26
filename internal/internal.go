package internal

import "flag"

var (
	PackageDestDir = flag.String("package-dest-dir", "./var", "directory to store built packages")
)
