//go:build windows

package diskinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/yusufpapurcu/wmi"
	"os/exec"
)

type Win32_DiskDrive struct {
	DeviceID  string
	MediaType string
	Status    string
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

// MSStorageDriver_FailurePredictData structure for SMART data
type MSStorageDriver_FailurePredictData struct {
	InstanceName   string
	VendorSpecific []byte
}

func parseTemperature(vendorSpecific []byte) int {
	if len(vendorSpecific) < 5 {
		return -1 // Invalid data
	}
	// Example: Extract temperature from SMART attribute ID 0xC2 (194)
	for i := 0; i < len(vendorSpecific)-6; i += 12 {
		if vendorSpecific[i] == 0xC2 {
			return int(vendorSpecific[i+5]) // Temperature value
		}
	}
	return -1 // Temperature attribute not found
}

type DiskInfo2 struct {
	DeviceID          string `json:"DeviceID"`
	MediaType         string `json:"MediaType"`
	HealthStatus      string `json:"HealthStatus"`
	OperationalStatus string `json:"OperationalStatus"`
	Temperature       int    `json:"Temperature"`
}

func tttt() {
	powerShellCmd := `
	@(
	    Get-PhysicalDisk | ForEach-Object {
	        $disk = $_
	        $reliability = Get-StorageReliabilityCounter -PhysicalDisk $disk
	        [PSCustomObject]@{
	            DeviceID = $disk.DeviceID
	            MediaType = $disk.MediaType
	            HealthStatus = $disk.HealthStatus
	            OperationalStatus = $disk.OperationalStatus
	            Temperature = $reliability.Temperature
	        }
	    }
	) | ConvertTo-Json -Depth 2
	`
	fmt.Println("Running command") // Debugging raw output

	// Execute the PowerShell command
	cmd := exec.Command("powershell", "-Command", powerShellCmd)

	output, err := cmd.CombinedOutput()
	fmt.Println("Raw Output:", string(output)) // Debugging raw output

	if err != nil {
		fmt.Printf("Error executing PowerShell command: %v\n", err)
		fmt.Println("Raw Output:", string(output)) // Debugging raw output

		return
	}

	// Parse the JSON output
	var disks []DiskInfo2
	err = json.Unmarshal(output, &disks)
	if err != nil {
		// If parsing as an array fails, try parsing as a single object
		var singleDisk DiskInfo2
		err = json.Unmarshal(output, &singleDisk)
		if err != nil {
			fmt.Printf("Error parsing JSON output: %v\n", err)
			return
		}
		// Wrap the single object in a slice
		disks = append(disks, singleDisk)
	}

	// Print the disk information
	for _, disk := range disks {
		fmt.Printf("DeviceID: %s\n", disk.DeviceID)
		fmt.Printf("MediaType: %s\n", disk.MediaType)
		fmt.Printf("HealthStatus: %s\n", disk.HealthStatus)
		fmt.Printf("OperationalStatus: %s\n", disk.OperationalStatus)
		fmt.Printf("Temperature: %d°C\n", disk.Temperature)
		fmt.Println("-------------------------")
	}
}

func getWin32_DiskDrive() ([]Win32_DiskDrive, map[string]MSStorageDriver_FailurePredictStatus, error) {
	tttt()
	return nil, nil, nil
	command := `
	Get-PhysicalDisk | ForEach-Object {
		$disk = $_
		$reliability = Get-StorageReliabilityCounter -PhysicalDisk $disk
		[PSCustomObject]@{
			DeviceID = $disk.DeviceID
			MediaType = $disk.MediaType
			OperationalStatus = $disk.OperationalStatus
			Temperature = $reliability.Temperature
		}
	} | ConvertTo-Json -Depth 2
	`
	cmd := exec.Command("powershell", "-Command", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		fmt.Println("Error:", err)
	}
	fmt.Println(string(output))

	var diskDrives []Win32_DiskDrive
	err = wmi.Query("SELECT DeviceID, SerialNumber, Model, Size, Status FROM Win32_DiskDrive", &diskDrives)
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
