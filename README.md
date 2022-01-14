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

func main() {
	f(map[string]int{})
	f(map[string]int{})
}

func f(m map[string]int) {
	println(len(m))
}
$ ptrls `pwd`/a.go 80
m
	 a.go:4:18 map[string]int
	 a.go:5:18 map[string]int
```
