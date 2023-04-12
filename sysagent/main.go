package main

import (
	"context"
	"fmt"
	influxdb2 "github.com/influxdata/influxdb-client-go/v2"
	"github.com/influxdata/influxdb-client-go/v2/api/write"
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/disk"
	"github.com/shirou/gopsutil/mem"
	"github.com/shirou/gopsutil/net"
	"github.com/sirupsen/logrus"
	"time"
)

var (
	url    = "http://127.0.0.1:8086"
	token  = "OoYd_khzbkGUaEsrbnwA2K2gg2kaT6gEVRVQ0itDkawRXv53AVY5NQuDczF-hYFRpWvHY2NdlZ9q50h39cOHjg=="
	org    = "my_org"
	bucket = "log_monitor"
	Client influxdb2.Client
)

func initConnInfluxDB() {
	// Create client
	Client = influxdb2.NewClient(url, token)
	defer Client.Close()
}

func writePoint(pt Point) {
	writeAPI := Client.WriteAPIBlocking(org, bucket)
	point := write.NewPoint(pt.measurement, pt.tags, pt.fields, time.Now())
	if err := writeAPI.WritePoint(context.Background(), point); err != nil {
		logrus.Errorf("write to influxdb failed, err: %v\n", err)
	}
}
func getCpuInfo() {
	// CPU使用率
	percent, _ := cpu.Percent(time.Second, false)
	fmt.Printf("cpu percent:%v\n", percent)
	if len(percent) == 0 {
		return
	}
	tags := map[string]string{
		"cpu": "cpu0",
	}
	fields := map[string]interface{}{
		"cpu_percent": percent[0],
	}
	writePoint(Point{
		tags:        tags,
		fields:      fields,
		measurement: "cpu",
	})
}
func getMemInfo() {
	memInfo, _ := mem.VirtualMemory()
	fmt.Printf("mem info:%v\n", memInfo)

	tags := map[string]string{
		"mem": "mem",
	}
	fields := map[string]interface{}{
		"total":        memInfo.Total,
		"available":    memInfo.Available,
		"used":         memInfo.Used,
		"used_percent": memInfo.UsedPercent,
		"buffers":      memInfo.Buffers,
		"cached":       memInfo.Cached,
	}
	writePoint(Point{
		tags:        tags,
		fields:      fields,
		measurement: "memory",
	})
}
func getDiskInfo() {
	parts, err := disk.Partitions(true)
	if err != nil {
		logrus.Errorf("get Partitions failed, err:%v\n", err)
		return
	}
	for _, part := range parts {
		logrus.Info("part:%v\n", part.String())
		diskInfo, err := disk.Usage(part.Mountpoint)
		if err != nil {
			logrus.Errorf("get usage stats failed, err: %v\n", err)
			continue
		}
		tags := map[string]string{
			"disk": "disk",
		}
		fields := map[string]interface{}{
			"path":              diskInfo.Path,
			"fstype":            diskInfo.Fstype,
			"total":             diskInfo.Total,
			"free":              diskInfo.Free,
			"used":              diskInfo.Used,
			"usedPercent":       diskInfo.UsedPercent,
			"inodesTotal":       diskInfo.InodesTotal,
			"inodesUsed":        diskInfo.InodesUsed,
			"inodesFree":        diskInfo.InodesFree,
			"inodesUsedPercent": diskInfo.InodesUsedPercent,
		}
		writePoint(Point{
			tags:        tags,
			fields:      fields,
			measurement: part.Mountpoint,
		})
	}
}
func getNetInfo() {
	info, err := net.IOCounters(true)
	if err != nil {
		logrus.Errorf("net get IOCounters failed, err: %v\n", err)
		return
	}
	for index, v := range info {
		logrus.Info("%v:%v send:%v recv:%v\n", index, v, v.BytesSent, v.BytesRecv)

		tags := map[string]string{
			"net": "net",
		}
		fields := map[string]interface{}{
			"bytesSent":   v.BytesSent,
			"bytesRecv":   v.BytesRecv,
			"packetsSent": v.PacketsSent,
			"packetsRecv": v.PacketsRecv,
			"errin":       v.Errin,
			"errout":      v.Errout,
			"dropin":      v.Dropin,
			"dropout":     v.Dropout,
			"fifoin":      v.Fifoin,
			"fifoout":     v.Fifoout,
		}
		writePoint(Point{
			tags:        tags,
			fields:      fields,
			measurement: v.Name,
		})
	}
}
func main() {
	initConnInfluxDB()

	//每1s执行一次
	ticker := time.NewTicker(time.Second)
	for range ticker.C {
		getCpuInfo()
		getMemInfo()
		getDiskInfo()
		getNetInfo()
	}
}
