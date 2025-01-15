//go:build windows

package diskinfo

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"os/exec"
)

type HealthStatusType string

const (
	HealthStatusHealthy   HealthStatusType = "Healthy"
	HealthStatusUnhealthy HealthStatusType = "Unhealthy"
	HealthStatusWarning   HealthStatusType = "Warning"
	HealthStatusUnknown   HealthStatusType = "Unknown"
)

type Win32Volume struct {
	DeviceID          string           `json:"DeviceID"`
	MediaType         string           `json:"MediaType"`
	HealthStatus      HealthStatusType `json:"HealthStatus"`
	OperationalStatus string           `json:"OperationalStatus"`
	SerialNumber      string           `json:"SerialNumber"`
	Model             string           `json:"Model"`
	FirmwareVersion   string           `json:"FirmwareVersion"`
	Size              int64            `json:"Size"`
	Temperature       int              `json:"Temperature"`
	ReadErrors        interface{}      `json:"ReadErrors"`
	WriteErrors       interface{}      `json:"WriteErrors"`
	PowerOnHours      interface{}      `json:"PowerOnHours"`
	UsedSpace         int              `json:"UsedSpace"`
	FreeSpace         int              `json:"FreeSpace"`
}

func getWin32_Volume() ([]Win32Volume, error) {
	command := `
	@(
    Get-PhysicalDisk | ForEach-Object {
        $disk = $_
        $reliability = Get-StorageReliabilityCounter -PhysicalDisk $disk
        $volumes = Get-WmiObject -Query "SELECT FreeSpace, Capacity, DriveLetter FROM Win32_Volume WHERE DriveLetter IS NOT NULL" | Where-Object {
            $_.DriveLetter -eq ($disk.DeviceID -replace '\\\.\\', '')
        }
        
        [PSCustomObject]@{
            DeviceID = $disk.DeviceID
            MediaType = $disk.MediaType
            HealthStatus = $disk.HealthStatus
            OperationalStatus = $disk.OperationalStatus
            SerialNumber = $disk.SerialNumber
            Model = $disk.Model
            FirmwareVersion = $disk.FirmwareVersion
            Size = $disk.Size
            Temperature = $reliability.Temperature
            ReadErrors = $reliability.ReadErrorsTotal
            WriteErrors = $reliability.WriteErrorsTotal
            PowerOnHours = $reliability.PowerOnHours
            UsedSpace = $volumes | ForEach-Object { $_.Capacity - $_.FreeSpace } | Measure-Object -Sum | Select-Object -ExpandProperty Sum
            FreeSpace = $volumes | ForEach-Object { $_.FreeSpace } | Measure-Object -Sum | Select-Object -ExpandProperty Sum
        }
			}
		) | ConvertTo-Json -Depth 2 -Compress
	`
	logrus.Debugf("Running commend: powershell -Command %s", command)
	cmd := exec.Command("powershell", "-Command", command)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("error running powershell: %v", err)
	}
	var volumes []Win32Volume
	err = json.Unmarshal(output, &volumes)
	if err != nil {
		var singleDisk Win32Volume
		err = json.Unmarshal(output, &singleDisk)
		if err != nil {
			return nil, fmt.Errorf("Error parsing JSON output: %v", err)
		}
		volumes = append(volumes, singleDisk)
	}
	return volumes, nil
}

type WindowsDiskInfo struct {
	ctx context.Context
}

func getStatus(healthStatus HealthStatusType) (*StatusType, error) {
	var status StatusType
	switch healthStatus {
	case HealthStatusUnhealthy:
		status = StatusError
	case HealthStatusUnknown:

		status = StatusUnknown
	case HealthStatusHealthy:

		status = StatusSafe
	case HealthStatusWarning:

		status = StatusWarning
	default:
		return nil, fmt.Errorf("Invalid HealthStatusType: %s", healthStatus)
	}
	return &status, nil
}

func transform(w Win32Volume) DiskInfo {
	status, _ := getStatus(w.HealthStatus)
	return DiskInfo{
		Status:      *status,
		Condition:   "",
		DeviceName:  w.DeviceID,
		Temperature: w.Temperature,
	}
}

func (l WindowsDiskInfo) GetDisksInfo() ([]DiskInfo, error) {
	var disksInfos []DiskInfo
	volumes, err := getWin32_Volume()
	if err != nil {
		return nil, fmt.Errorf("failed to revtrieve disk data: %v", err)
	}
	for _, volume := range volumes {
		disksInfos = append(disksInfos, transform(volume))
	}
	return disksInfos, nil
}

func NewDiskInfoProvider(ctx context.Context) DiskInfoProvider {
	return WindowsDiskInfo{ctx: ctx}
}
