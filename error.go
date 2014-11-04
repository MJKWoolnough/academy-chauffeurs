// +build !windows, !linux

package main

import (
	"fmt"
	"os"
)

func ShowError(err error) {
	fmt.Fprintln(os.Stderr, err)
}
