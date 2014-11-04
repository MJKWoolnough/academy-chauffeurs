package main

import "github.com/AllenDang/w32"

func ShowError(err error) {
	w32.MessageBox(0, "Error", err.Error(), w32.MB_OK|w32.MB_ICONERROR)
}
