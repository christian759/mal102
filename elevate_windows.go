//go:build windows
// +build windows

package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"unsafe"

	"golang.org/x/sys/windows"
)

func IsElevated() (bool, error) {
	var token windows.Token
	err := windows.OpenProcessToken(windows.CurrentProcess(), windows.TOKEN_QUERY, &token)
	if err != nil {
		return false, err
	}
	defer token.Close()

	var elevation uint32
	var returned uint32
	err = windows.GetTokenInformation(token, windows.TokenElevation, (*byte)(unsafe.Pointer(&elevation)), uint32(unsafe.Sizeof(elevation)), &returned)
	if err != nil {
		return false, err
	}
	return elevation != 0, nil
}

func RelaunchElevated() error {
	// Build quoted arg list for PowerShell -ArgumentList
	args := ""
	if len(os.Args) > 1 {
		// naive join; if you need perfect escaping consider more care
		args = `"` + strings.Join(os.Args[1:], `" "`) + `"`
	}
	psCmd := fmt.Sprintf("Start-Process -FilePath '%s' -ArgumentList %s -Verb RunAs", os.Args[0], args)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
