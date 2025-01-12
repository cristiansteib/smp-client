package main

import (
	"fmt"
	"os"
	"runtime"
)

func isAdmin() (bool, error) {
	switch runtime.GOOS {
	case "windows":
		// TODO
		return true, nil
	case "linux":
		return os.Geteuid() == 0, nil
	default:
		return false, fmt.Errorf("unsupported platform")
	}
}

func assertIsRunningWithRequiredPrivileges() {
	isAdmin, err := isAdmin()
	if err != nil {
		fmt.Printf("Error checking admin status: %v\n", err)
		os.Exit(1)
	}
	if !isAdmin {
		fmt.Println("Program is NOT running with administrator privileges.")
		os.Exit(2)
	}
}
