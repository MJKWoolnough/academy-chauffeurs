package main

import "os/exec"

func OpenBrowser(url string) *exec.Cmd {
	return exec.Command("open", url)
}
