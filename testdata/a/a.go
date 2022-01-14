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
