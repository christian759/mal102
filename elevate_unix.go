//go:build linux || darwin
// +build linux darwin

package main

import (
	"fmt"
	"os"
	"os/exec"
	"syscall"
)

func IsElevated() (bool, error) {
	return syscall.Geteuid() == 0, nil
}

func RelaunchElevated() error {
	args := append([]string{os.Args[0]}, os.Args[1:]...)
	cmd := exec.Command("sudo", args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	fmt.Println("Requesting sudo elevation...")
	return cmd.Run()
}
