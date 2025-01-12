package diskinfo

type StatusType string

const (
	StatusSafe    StatusType = "Safe"
	StatusWarning StatusType = "Warning"
	StatusError   StatusType = "Error"
)

type DiskInfo struct {
	Status      StatusType
	Condition   string
	DeviceName  string
	Temperature int
}

func (i DiskInfo) StatusToInt() int {
	switch i.Status {
	case StatusSafe:
		return 0
	case StatusWarning:
		return 1
	case StatusError:
		return 2
	default:
		return -1
	}
}

type DiskInfoProvider interface {
	GetDisksInfo() ([]DiskInfo, error)
}
