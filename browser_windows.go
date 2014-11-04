package main

import "os/exec"

func OpenBrowser(url string) *exec.Cmd {
	return exec.Command("rundll32", "url.dll", "FileProtocolHandler", url)
}
