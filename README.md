# ptrls

## Install

```
$ go install github.com/gostaticanalysis/ptrls/cmd/ptrls@latest
```

## Usage

```
$ cd testdata/a
$ cat a.go
package main

import (
	"bytes"
	"fmt"
	"io"
	"os"
)

func main() {
	f(os.Stdout)
	f(new(bytes.Buffer))
}

func f(w io.Writer) {
	fmt.Println(w, "hello")
}
$ ptrls `pwd`/a.go 114 
w
	 makeinterface:*os.File
	 makeinterface:*bytes.Buffer
```
