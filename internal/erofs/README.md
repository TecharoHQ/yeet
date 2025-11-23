# erofs

This package was inlined from [`github.com/dpeckett/archivefs`' subpackage `erofs`](https://github.com/dpeckett/archivefs/tree/main) with one crucial difference: all file timestamps are set to unix time one.

# archivefs

Implementations of Go's [io/fs.FS](https://pkg.go.dev/io/fs#FS) interface for
various archive types.

## Supported Archive Types

- [ar](<https://en.wikipedia.org/wiki/Ar_(Unix)>)
- [erofs](https://en.wikipedia.org/wiki/EROFS)
- [tar](<https://en.wikipedia.org/wiki/Tar_(computing)>)

## Usage

```go
package main

import (
  "log"
  "os"

  "github.com/dpeckett/archivefs/tarfs"
)

func main() {
  f, err := os.Open("example.tar")
  if err != nil {
    log.Fatal(err)
  }
  defer f.Close()

  fsys, err := tarfs.Open(f)
  if err != nil {
    log.Fatal(err)
  }

  // Do something with the filesystem.
}
```

## License

This project is licensed under the Mozilla Public License 2.0 - see the
[LICENSE](LICENSE) file for details.
