//go:build darwin

package main

import (
	"github.com/s-christian/gollehs/lib/system"
)

const (
	// Shell constants
	bash = "/bin/bash"
	sh   = "/bin/sh"
)

func GetShellName() (shell, args string) {
	args = "-c"

	if system.FileExists(bash) {
		shell = bash
	} else {
		shell = sh
	}

	return
}
