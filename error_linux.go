package main

import (
	"fmt"
	"os"
	"os/exec"
)

func ShowError(err error) {
	if e := exec.Command("notify-send", err.Error()).Run(); e != nil {
		fmt.Fprintln(os.Stderr, err)
	}
}
