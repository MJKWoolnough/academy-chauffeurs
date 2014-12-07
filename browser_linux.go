package main

import (
	"github.com/CzarekTomczak/cef2go/src/cef"
	"github.com/CzarekTomczak/cef2go/src/gtk"
)

func OpenBrowser(url string) {
	cef.ExecuteProcess(nil)
	cef.Initialize(cef.Settings{})
	gtk.Initialize()
	window := gtk.CreateWindow("AC", 1024, 768)
	gtk.ConnectDestroySignal(window, func() {
		cef.QuitMessageLoop()
	})
	cef.CreateBrowser(window, cef.BrowserSettings{}, url)
	cef.RunMessageLoop()
	cef.Shutdown()
}
