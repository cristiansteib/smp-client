package internal

import (
	"context"
	"gama-client/internal/appconfig"
	"gama-client/internal/diskinfo"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/sirupsen/logrus"
	"time"
)

func sendDiskInfo(ctx context.Context, cancelFunc context.CancelFunc, writeAPI api.WriteAPIBlocking, diskInfoProvider diskinfo.DiskInfoProvider) {
	disks, err := diskInfoProvider.GetDisksInfo()
	if err != nil {
		logrus.Errorf("Failed to retrieve disk info: %v", err)
		time.Sleep(30 * time.Second)
		cancelFunc()
	}
	for _, diskInfo := range disks {
		tags := map[string]string{
			"host":   "hostName",
			"client": "escuela1",
			"device": diskInfo.DeviceName,
		}
		fields := map[string]interface{}{
			"status":      diskInfo.StatusToInt(),
			"temperature": diskInfo.Temperature,
		}
		point := write.NewPoint("disk", tags, fields, time.Now())
		time.Sleep(5 * time.Second)

		if err := writeAPI.WritePoint(ctx, point); err != nil {
			logrus.Errorf("Error sending flux point %v", err)
		}
	}
}

func Service(ctx context.Context, cancelFunc context.CancelFunc, config *appconfig.AppConfig) {
	client := influxdb2.NewClient(config.InfluxURL, config.InfluxToken)
	defer client.Close()
	writeAPI := client.WriteAPIBlocking(config.InfluxOrg, config.InfluxBucket)
	diskInfoProvider := diskinfo.NewDiskInfoProvider(ctx)
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			cancelFunc()
			return
		case <-ticker.C:
			sendDiskInfo(ctx, cancelFunc, writeAPI, diskInfoProvider)
		}
	}
}
