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

// RelaunchElevated uses PowerShell Start-Process -Verb RunAs to trigger UAC.
// It builds a quoted argument list (naive quoting) — this is fine for most simple cases.
func RelaunchElevated() error {
	args := ""
	if len(os.Args) > 1 {
		// simple join with quoting; adapt if you need complex argument escaping
		quoted := make([]string, 0, len(os.Args)-1)
		for _, a := range os.Args[1:] {
			// escape single quotes inside argument
			safe := strings.ReplaceAll(a, "'", "''")
			quoted = append(quoted, "'"+safe+"'")
		}
		args = strings.Join(quoted, ",")
		args = "-ArgumentList " + args
	}
	// Build the PowerShell command
	psCmd := fmt.Sprintf("Start-Process -FilePath '%s' %s -Verb RunAs", os.Args[0], args)
	cmd := exec.Command("powershell", "-NoProfile", "-Command", psCmd)
	// Let the new elevated process interact with the user via UAC
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
