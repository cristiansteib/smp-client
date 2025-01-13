//go:build linux

package adminprivileges

import "os"

type LinuxPrivileges struct{}

func NewAdminPrivileges() AdminPrivileges {
	return &LinuxPrivileges{}
}

func (a *LinuxPrivileges) Check() (bool, error) {
	return os.Geteuid() == 0, nil
}
