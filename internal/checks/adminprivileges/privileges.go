package adminprivileges

type AdminPrivileges interface {
	Check() (bool, error)
}
