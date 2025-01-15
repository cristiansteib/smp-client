//go:build windows

package adminprivileges

type WindowsPrivileges struct{}

func NewAdminPrivileges() AdminPrivileges {
	return &WindowsPrivileges{}
}

func (a *WindowsPrivileges) Check() (bool, error) {
	return true, nil
}
