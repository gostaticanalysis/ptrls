package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"

	"github.com/gostaticanalysis/ptrls"
)

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, "ptrls:", err)
		os.Exit(1)
	}
}

func run(args []string) error {
	if len(args) != 2 {
		return errors.New("file name and offset must be specified")
	}

	prog, err := ptrls.Load("./...")
	if err != nil {
		return err
	}

	offset, err := strconv.Atoi(args[1])
	if err != nil {
		return err
	}

	pos := prog.Pos(args[0], offset)

	ptrs, err := ptrls.Analyze(prog, pos)
	if err != nil {
		return err
	}

	for n, ptr := range ptrs {
		fmt.Println(n)
		for _, l := range ptr.PointsTo().Labels() {
			fmt.Println("\t", l)
		}
		fmt.Println()
	}

	return nil
}
