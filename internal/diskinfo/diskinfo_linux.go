//go:build linux

package diskinfo

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/sirupsen/logrus"
	"log"
	"os/exec"
	"strings"
)

type SmartctlOutput struct {
	JSONFormatVersion             []int               `json:"json_format_version"`
	NVMESMARTHealthInformationLog *NVMELog            `json:"nvme_smart_health_information_log,omitempty"`
	Smartctl                      SmartctlInfo        `json:"smartctl"`
	Device                        DeviceInfo          `json:"device"`
	ModelFamily                   string              `json:"model_family,omitempty"`
	ModelName                     string              `json:"model_name,omitempty"`
	SerialNumber                  string              `json:"serial_number,omitempty"`
	WWN                           *WWN                `json:"wwn,omitempty"`
	FirmwareVersion               string              `json:"firmware_version,omitempty"`
	UserCapacity                  *Capacity           `json:"user_capacity,omitempty"`
	LogicalBlockSize              *int                `json:"logical_block_size,omitempty"`
	PhysicalBlockSize             *int                `json:"physical_block_size,omitempty"`
	RotationRate                  *int                `json:"rotation_rate,omitempty"`
	InSmartctlDatabase            bool                `json:"in_smartctl_database,omitempty"`
	ATAVersion                    *ATAVersion         `json:"ata_version,omitempty"`
	SATAVersion                   *SATAVersion        `json:"sata_version,omitempty"`
	InterfaceSpeed                *InterfaceSpeed     `json:"interface_speed,omitempty"`
	LocalTime                     LocalTime           `json:"local_time"`
	ReadLookahead                 *FeatureStatus      `json:"read_lookahead,omitempty"`
	WriteCache                    *FeatureStatus      `json:"write_cache,omitempty"`
	ATASecurity                   *ATASecurity        `json:"ata_security,omitempty"`
	SmartStatus                   SmartStatus         `json:"smart_status"`
	ATASMARTData                  *ATASMARTData       `json:"ata_smart_data,omitempty"`
	ATASMARTAttributes            *ATASMARTAttributes `json:"ata_smart_attributes,omitempty"`
	Temperature                   *Temperature        `json:"temperature,omitempty"`
	PowerCycleCount               *int                `json:"power_cycle_count,omitempty"`
	PowerOnTime                   *PowerOnTime        `json:"power_on_time,omitempty"`
}

type NVMELog struct {
	CriticalWarning         int `json:"critical_warning"`
	AvailableSpare          int `json:"available_spare"`
	AvailableSpareThreshold int `json:"available_spare_threshold"`
	PercentageUsed          int `json:"percentage_used"`
	MediaErrors             int `json:"media_errors"`
	NumErrLogEntries        int `json:"num_err_log_entries"`
}

// Structs for nested fields
type SmartctlInfo struct {
	Version      []int    `json:"version"`
	SVNRevision  string   `json:"svn_revision"`
	PlatformInfo string   `json:"platform_info"`
	BuildInfo    string   `json:"build_info"`
	Argv         []string `json:"argv"`
	ExitStatus   int      `json:"exit_status"`
}

type DeviceInfo struct {
	Name     string `json:"name"`
	InfoName string `json:"info_name"`
	Type     string `json:"type"`
	Protocol string `json:"protocol"`
}

type NVMePCIVendor struct {
	ID          int `json:"id"`
	SubsystemID int `json:"subsystem_id"`
}

type NVMENamespace struct {
	ID               int      `json:"id"`
	Size             Capacity `json:"size"`
	Capacity         Capacity `json:"capacity"`
	Utilization      Capacity `json:"utilization"`
	FormattedLBASize int      `json:"formatted_lba_size"`
	EUI64            *EUI64   `json:"eui64,omitempty"`
}

type Capacity struct {
	Blocks int64 `json:"blocks"`
	Bytes  int64 `json:"bytes"`
}

type EUI64 struct {
	OUI   int   `json:"oui"`
	ExtID int64 `json:"ext_id"`
}

type LocalTime struct {
	TimeT   int    `json:"time_t"`
	Asctime string `json:"asctime"`
}

type SmartStatus struct {
	Passed bool      `json:"passed"`
	NVMe   *NVMeInfo `json:"nvme,omitempty"`
}

type NVMeInfo struct {
	Value int `json:"value"`
}

type NVMESMARTHealthInfoLog struct {
	CriticalWarning         int   `json:"critical_warning"`
	Temperature             int   `json:"temperature"`
	AvailableSpare          int   `json:"available_spare"`
	AvailableSpareThreshold int   `json:"available_spare_threshold"`
	PercentageUsed          int   `json:"percentage_used"`
	DataUnitsRead           int64 `json:"data_units_read"`
	DataUnitsWritten        int64 `json:"data_units_written"`
	HostReads               int64 `json:"host_reads"`
	HostWrites              int64 `json:"host_writes"`
	ControllerBusyTime      int64 `json:"controller_busy_time"`
	PowerCycles             int   `json:"power_cycles"`
	PowerOnHours            int   `json:"power_on_hours"`
	UnsafeShutdowns         int   `json:"unsafe_shutdowns"`
	MediaErrors             int   `json:"media_errors"`
	NumErrLogEntries        int   `json:"num_err_log_entries"`
	WarningTempTime         int   `json:"warning_temp_time"`
	CriticalCompTime        int   `json:"critical_comp_time"`
}

type Temperature struct {
	Current int `json:"current"`
}

type PowerOnTime struct {
	Hours int `json:"hours"`
}

type WWN struct {
	NAA int   `json:"naa"`
	OUI int   `json:"oui"`
	ID  int64 `json:"id"`
}

type ATAVersion struct {
	String     string `json:"string"`
	MajorValue int    `json:"major_value"`
	MinorValue int    `json:"minor_value"`
}

type SATAVersion struct {
	String string `json:"string"`
	Value  int    `json:"value"`
}

type InterfaceSpeed struct {
	Max     SpeedInfo `json:"max"`
	Current SpeedInfo `json:"current"`
}

type SpeedInfo struct {
	SATAValue      int    `json:"sata_value"`
	String         string `json:"string"`
	UnitsPerSecond int    `json:"units_per_second"`
	BitsPerUnit    int    `json:"bits_per_unit"`
}

type FeatureStatus struct {
	Enabled bool `json:"enabled"`
}

type ATASecurity struct {
	State   int    `json:"state"`
	String  string `json:"string"`
	Enabled bool   `json:"enabled"`
	Frozen  bool   `json:"frozen"`
}

type ATASMARTData struct {
	OfflineDataCollection OfflineDataCollection `json:"offline_data_collection"`
	SelfTest              SelfTest              `json:"self_test"`
	Capabilities          Capabilities          `json:"capabilities"`
}

type OfflineDataCollection struct {
	Status            Status `json:"status"`
	CompletionSeconds int    `json:"completion_seconds"`
}

type Status struct {
	Value  int    `json:"value"`
	String string `json:"string"`
	Passed bool   `json:"passed"`
}

type SelfTest struct {
	Status         Status       `json:"status"`
	PollingMinutes PollingTimes `json:"polling_minutes"`
}

type PollingTimes struct {
	Short      int `json:"short"`
	Extended   int `json:"extended"`
	Conveyance int `json:"conveyance"`
}

type Capabilities struct {
	Values                        []int `json:"values"`
	ExecOfflineImmediateSupported bool  `json:"exec_offline_immediate_supported"`
	OfflineIsAbortedUponNewCmd    bool  `json:"offline_is_aborted_upon_new_cmd"`
	OfflineSurfaceScanSupported   bool  `json:"offline_surface_scan_supported"`
	SelfTestsSupported            bool  `json:"self_tests_supported"`
	ConveyanceSelfTestSupported   bool  `json:"conveyance_self_test_supported"`
	SelectiveSelfTestSupported    bool  `json:"selective_self_test_supported"`
	AttributeAutosaveEnabled      bool  `json:"attribute_autosave_enabled"`
	ErrorLoggingSupported         bool  `json:"error_logging_supported"`
	GPLoggingSupported            bool  `json:"gp_logging_supported"`
}

type ATASMARTAttributes struct {
	Revision int              `json:"revision"`
	Table    []SMARTAttribute `json:"table"`
}

type SMARTAttribute struct {
	ID         int        `json:"id"`
	Name       string     `json:"name"`
	Value      int        `json:"value"`
	Worst      int        `json:"worst"`
	Thresh     int        `json:"thresh"`
	WhenFailed string     `json:"when_failed"`
	Flags      SMARTFlags `json:"flags"`
	Raw        SMARTRaw   `json:"raw"`
}

type SMARTFlags struct {
	Value         int    `json:"value"`
	String        string `json:"string"`
	Prefailure    bool   `json:"prefailure"`
	UpdatedOnline bool   `json:"updated_online"`
	Performance   bool   `json:"performance"`
	ErrorRate     bool   `json:"error_rate"`
	EventCount    bool   `json:"event_count"`
	AutoKeep      bool   `json:"auto_keep"`
}

type SMARTRaw struct {
	Value  int    `json:"value"`
	String string `json:"string"`
}

func NewSmartData(raw []byte) (*SmartctlOutput, error) {
	var data SmartctlOutput
	if err := json.Unmarshal(raw, &data); err != nil {
		return nil, err
	}
	return &data, nil
}

func (sctl *SmartctlOutput) ClassifyDisk() (StatusType, string) {
	// 1. Verificar si el estado SMART global no es aceptable
	if sctl == nil {
		logrus.Fatalf("Error on SmartctlOutput::ClassifyDisk nil pointer")
	}

	if !sctl.SmartStatus.Passed {
		return StatusError, "SMART status check failed"
	}

	// 2. Verificar errores específicos para discos NVMe
	if sctl.NVMESMARTHealthInformationLog != nil {
		healthLog := sctl.NVMESMARTHealthInformationLog

		if healthLog.CriticalWarning != 0 {
			return StatusError, "Critical warning detected in NVMe log"
		}

		if healthLog.MediaErrors > 0 {
			return StatusError, "Media errors found in NVMe log"
		}

		if healthLog.AvailableSpare < healthLog.AvailableSpareThreshold {
			return StatusWarning, "Available spare below threshold in NVMe log"
		}

		if healthLog.PercentageUsed >= 80 {
			return StatusWarning, "Percentage used exceeds 80% in NVMe log"
		}

		if healthLog.NumErrLogEntries > 100 {
			return StatusWarning, "Excessive error log entries in NVMe log"
		}
	}

	// 3. Verificar errores específicos para discos ATA/SATA
	if sctl.ATASMARTAttributes != nil {
		attributes := sctl.ATASMARTAttributes.Table

		// Verificar atributos relevantes de sectores reasignados o pendientes
		for _, attr := range attributes {
			if attr.ID == 5 && attr.Raw.Value > 0 {
				return StatusWarning, "Reallocated sectors count is greater than 0"
			}
			if attr.ID == 196 && attr.Raw.Value > 0 {
				return StatusWarning, "Reallocated event count is greater than 0"
			}
			if attr.ID == 197 && attr.Raw.Value > 0 {
				return StatusWarning, "Current pending sector count is greater than 0"
			}
		}
	}

	// 4. Verificar temperatura
	if sctl.Temperature != nil && sctl.Temperature.Current > 70 {
		return StatusWarning, "Temperature exceeds 70°C"
	}

	// Si todas las verificaciones pasan, el estado es "Safe"
	return StatusSafe, "All checks passed"
}

type LinuxDiskInfo struct {
	ctx context.Context
}

func getSmartData(disk string) (*SmartctlOutput, error) {
	cmdR := fmt.Sprintf("%s %s \n", "sudo smartctl -a -x -j", disk)

	logrus.Debugf("Running: %s", cmdR)
	cmd := exec.Command("smartctl", "-a", "-x", "-j", disk)
	output, err := cmd.CombinedOutput()

	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			exitCode := exitErr.ExitCode()
			if exitCode == 1 || exitCode == 2 {
				logrus.Errorf("Failed to run smartctl command. [%s]", cmdR)
				return nil, err
			}
		} else {
			logrus.Fatalf("Unexpected error. [%v]", err)
			return nil, err
		}
	}
	data, err := NewSmartData(output)
	if err != nil {
		log.Printf("Error parsing smartctl: %v\n", err)
		return nil, err
	}
	return data, nil
}

func (l LinuxDiskInfo) GetDisksInfo() ([]DiskInfo, error) {
	fmt.Println("Fetching disk info on Linux using smartctl...")
	var disks []DiskInfo
	partitions, err := disk.PartitionsWithContext(l.ctx, false)
	if err != nil {
		return nil, fmt.Errorf("Error al obtener particiones: %v", err)
	}
	for _, partition := range partitions {
		device := partition.Device
		if strings.Contains(device, "snap") || strings.Contains(device, "loop") {
			continue
		}
		smartData, err := getSmartData(device)
		if err != nil {
			logrus.Errorf("Error retrieving device info: %s :: %v", device, err)
		}
		status, condition := smartData.ClassifyDisk()
		disks = append(disks, DiskInfo{
			Status:      status,
			Condition:   condition,
			DeviceName:  smartData.Device.Name,
			Temperature: smartData.Temperature.Current,
		})
	}

	return disks, nil
}

func NewDiskInfoProvider(ctx context.Context) DiskInfoProvider {
	return LinuxDiskInfo{ctx: ctx}
}
