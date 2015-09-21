package main

import (
	"bufio"
	"bytes"
	"errors"
	"io/ioutil"
	"log"
	"strconv"
	"strings"
	"time"

	"github.com/toorop/gopentsdb"
)

func main() {
	var diffUser, diffNice, diffSys, diffWait, diffIdle, diffSum uint64
	// Les variables suivant vont êtres utilisées pour stocker les stats entre
	// deux mesures
	var cStatsPrev map[string]CPUStats

	temporizedClient, err := gopentsdb.NewTemporizedClient(gopentsdb.TemporizedClientConfig{
		Period:    10,
		MaxPoints: 0,
		CConfig: &gopentsdb.ClientConfig{
			Endpoint:           "",
			Username:           "",
			Password:           "",
			InsecureSkipVerify: true,
		},
	})
	if err != nil {
		log.Fatalln(err)
	}

	//var tv syscall.Timeval
	for {
		time.Sleep(1000 * time.Millisecond)
		// Les stats cpu
		// cStatsPrev va stocker les stats cpu du test précédent. Ca va nous
		// permettre d'avoir des usage en % sur la periode donnée
		//syscall.Gettimeofday(&tv)
		//now := (int64(tv.Sec)*1e3 + int64(tv.Usec)/1e3)
		now := time.Now().UnixNano() / 1000000
		cStats, err := GetCPUStats()
		if err != nil {
			log.Print("ERROR: ", err)
			continue
		}

		// On génere les point pour tous les cpu/cores
		for cpu, stats := range *cStats {
			// si on a des stats antérieures on fait un dif pour connaitre la conso de
			// chaque en pourcentage sur l'interval
			if &cStatsPrev != nil {
				if prevStats, ok := cStatsPrev[cpu]; ok {
					diffUser = stats.User - prevStats.User
					diffNice = stats.Nice - prevStats.Nice
					diffSys = stats.Sys - prevStats.Sys
					diffWait = stats.Wait - prevStats.Wait
					diffIdle = stats.Idle - prevStats.Idle
					diffSum = diffUser + diffNice + diffSys + diffWait + diffIdle

					// user
					point := gopentsdb.NewPoint()
					point.Metric = "cpu." + cpu + ".user.percent"
					point.Timestamp = now
					point.Value = float64(diffUser * 100 / diffSum)
					point.Tags["host"] = "trooper"
					temporizedClient.Add(point)

					// nice
					point = gopentsdb.NewPoint()
					point.Metric = "cpu." + cpu + ".nice.percent"
					point.Timestamp = now
					point.Value = float64(diffNice * 100 / diffSum)
					point.Tags["host"] = "trooper"
					temporizedClient.Add(point)

					// sys
					point = gopentsdb.NewPoint()
					point.Metric = "cpu." + cpu + ".sys.percent"
					point.Timestamp = now
					point.Value = float64(diffSys * 100 / diffSum)
					point.Tags["host"] = "trooper"
					temporizedClient.Add(point)

					// wait
					point = gopentsdb.NewPoint()
					point.Metric = "cpu." + cpu + ".wait.percent"
					point.Timestamp = now
					point.Value = float64(diffWait * 100 / diffSum)
					point.Tags["host"] = "trooper"
					temporizedClient.Add(point)

					// idle
					point = gopentsdb.NewPoint()
					point.Metric = "cpu." + cpu + ".idle.percent"
					point.Timestamp = now
					point.Value = float64(diffIdle * 100 / diffSum)
					point.Tags["host"] = "trooper"
					temporizedClient.Add(point)
				}
			}
		}
		cStatsPrev = *cStats

	}

}

// CPUStats represente les stats CPU
type CPUStats struct {
	User    uint64
	Nice    uint64
	Sys     uint64
	Idle    uint64
	Wait    uint64
	Irq     uint64
	SoftIrq uint64
	Stolen  uint64
}

// Sum adds all "stats" (cpu time)
func (s *CPUStats) Sum() uint64 {
	return s.Idle + s.Irq + s.Nice + s.SoftIrq + s.Stolen + s.Sys + s.User + s.Wait
}

// GetCPUStats retourne les stats CPU
func GetCPUStats() (*map[string]CPUStats, error) {
	procStats, err := ioutil.ReadFile("/proc/stat")
	if err != nil {
		return nil, err
	}
	scanner := bufio.NewScanner(bytes.NewReader(procStats))
	cStats := make(map[string]CPUStats)
	for scanner.Scan() {
		line := scanner.Text()
		if !strings.HasPrefix(line, "cpu") {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) < 8 {
			return nil, errors.New("bad data found in /proc/stat - " + line)
		}
		cStats[parts[0]] = CPUStats{
			User:    parseUint64(parts[1]),
			Nice:    parseUint64(parts[2]),
			Sys:     parseUint64(parts[3]),
			Idle:    parseUint64(parts[4]),
			Wait:    parseUint64(parts[5]),
			Irq:     parseUint64(parts[6]),
			SoftIrq: parseUint64(parts[7]),
			Stolen:  parseUint64(parts[8]),
		}
	}
	return &cStats, nil
}

func parseUint64(in string) uint64 {
	out, _ := strconv.ParseUint(in, 10, 64)
	return out
}
