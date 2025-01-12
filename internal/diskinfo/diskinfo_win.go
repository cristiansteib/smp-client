//go:build windows

package diskinfo

import (
	"context"
	"fmt"
	"github.com/StackExchange/wmi"
	"github.com/sirupsen/logrus"
)

type Win32_DiskDrive struct {
	DeviceID     string
	SerialNumber string
	Model        string
	Size         uint64
	Status       string
}

type MSStorageDriver_FailurePredictStatus struct {
	InstanceName   string
	PredictFailure bool
}

func testWmiAccess() error {
	var testQuery []struct{}
	err := wmi.Query("SELECT * FROM Win32_OperatingSystem", &testQuery)
	if err != nil {
		return fmt.Errorf("WMI access validation failed: %v", err)
	}
	logrus.Info("WMI access validated successfully.")
	return nil
}

func getWin32_DiskDrive() ([]Win32_DiskDrive, map[string]MSStorageDriver_FailurePredictStatus, error) {

	var diskDrives []Win32_DiskDrive
	err := wmi.Query("SELECT DeviceID, SerialNumber, Model, Size, Status FROM Win32_DiskDrive", &diskDrives)
	if err != nil {
		logrus.Errorf("error querying WMI: %v \n", err)
		return nil, nil, fmt.Errorf("error querying WMI: %v", err)
	}

	var failureStatuses []MSStorageDriver_FailurePredictStatus
	err = wmi.Query("SELECT InstanceName, PredictFailure FROM MSStorageDriver_FailurePredictStatus", &failureStatuses)
	if err != nil {
		logrus.Errorf("error querying SMART status: %v \n", err)
		return nil, nil, fmt.Errorf("error querying SMART status: %v", err)
	}

	// Mapped by instanceName
	var failureStatusesMapped map[string]MSStorageDriver_FailurePredictStatus
	for _, status := range failureStatuses {
		failureStatusesMapped[status.InstanceName] = status
	}
	return diskDrives, failureStatusesMapped, nil
}

type WindowsDiskInfo struct {
	ctx context.Context
}

func (l WindowsDiskInfo) GetDisksInfo() ([]DiskInfo, error) {
	fmt.Println("Fetching disk info on Windows using WMI...")
	err := testWmiAccess()
	if err != nil {
		return nil, fmt.Errorf("access WMI Test failed: %v", err)
	}

	var disksInfo []DiskInfo
	disks, disksStatus, err := getWin32_DiskDrive()
	if err != nil {
		return nil, fmt.Errorf("failed to revtrieve disk data: %v", err)
	}

	for _, disk := range disks {
		var status StatusType
		fmt.Printf("%s \n", disk.DeviceID)

		if diskStatus, exists := disksStatus[disk.DeviceID]; exists {
			if diskStatus.PredictFailure {
				fmt.Println("  SMART Status: Warning - Predicted Failure")
				status = StatusWarning
			} else {
				status = StatusSafe
				fmt.Println("  SMART Status: OK")
			}
		}
		disksInfo = append(disksInfo, DiskInfo{
			Status:      status,
			Condition:   "",
			DeviceName:  disk.DeviceID,
			Temperature: 0,
		})
	}
	return nil, nil
}

func NewDiskInfoProvider(ctx context.Context) DiskInfoProvider {
	return WindowsDiskInfo{ctx: ctx}
}
