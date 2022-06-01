package slurmjob

import (
	"fmt"
	"log"
	"math"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

func parseTime(input string) (float64, uint64, uint64, uint64) {
	reg := `^(((?P<days>\d+)-)?(?P<hours>\d\d):)?(?P<minutes>\d\d):(?P<seconds>\d\d(\.\d+)?)$`
	r := regexp.MustCompile(reg)
	matches := r.FindStringSubmatch(input)
	var ss float64
	var mm, hh, dd uint64
	if matches != nil {
		ss, _ = strconv.ParseFloat(matches[r.SubexpIndex("seconds")], 64)
		mm, _ = strconv.ParseUint(matches[r.SubexpIndex("minutes")], 10, 32)
		hh, _ = strconv.ParseUint(matches[r.SubexpIndex("hours")], 10, 32)
		dd, _ = strconv.ParseUint(matches[r.SubexpIndex("days")], 10, 32)
	}
	return ss, mm, hh, dd
}

func parseByteSize(input string) uint64 {
	if input == "" || input == "16?" {
		return 0.0
	}
	m := map[string]float64{"K": 10, "M": 20, "G": 30, "T": 40, "P": 50, "E": 60}
	var value = 0.0
	var scale = 1.0
	if exp, found := m[input[len(input)-1:]]; found {
		scale = math.Pow(2, exp)
		value, _ = strconv.ParseFloat(input[:len(input)-1], 64)
	} else {
		value, _ = strconv.ParseFloat(input, 64)
	}
	return uint64(value * scale)
}

func parseCpuTime(input string) float64 {
	ss, mm, hh, dd := parseTime(input)
	return float64(dd*24*60*60+hh*60*60+mm*60) + ss
}

func ParseSstatMetrics(input []byte) (*SstatMetrics, error) {
	var metrics SstatMetrics
	lines := strings.Split(string(input), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		split := strings.Split(line, "|")
		if split[0] != "" {
			maxRSS := parseByteSize(split[1])
			if metrics.MaxRSS < maxRSS {
				metrics.MaxRSS = maxRSS
			}
		}
		if split[1] != "" {
			maxDiskWrite := parseByteSize(split[2])
			if metrics.MaxDiskWrite < maxDiskWrite {
				metrics.MaxDiskWrite = maxDiskWrite
			}
		}
		if split[2] != "" {
			maxDiskRead := parseByteSize(split[3])
			if metrics.MaxDiskRead < maxDiskRead {
				metrics.MaxDiskRead = maxDiskRead
			}
		}

	}
	return &metrics, nil
}

func ParseSacctMetrics(input []byte, l *log.Logger) (*SacctMetrics, error) {
	var metrics SacctMetrics
	if len(input) == 0 {
		return &metrics, nil
	}
	lines := strings.Split(string(input), "\n")
	for _, line := range lines {
		if line == "" {
			continue
		}
		split := strings.Split(line, "|")
		ncpus, _ := strconv.ParseInt(strings.TrimSpace(split[5]), 10, 16)
		if metrics.Ncpus < ncpus {
			metrics.Ncpus = ncpus
		}
		if split[17] != "" {
			maxRSS := parseByteSize(split[17])
			if metrics.MaxRSS < maxRSS {
				metrics.MaxRSS = maxRSS
			}
		}
		if split[18] != "" {
			maxDiskWrite := parseByteSize(split[18])
			if metrics.MaxDiskWrite < maxDiskWrite {
				metrics.MaxDiskWrite = maxDiskWrite
			}
		}
		if split[19] != "" {
			maxDiskRead := parseByteSize(split[19])
			if metrics.MaxDiskRead < maxDiskRead {
				metrics.MaxDiskRead = maxDiskRead
			}
		}

	}
	// retrieve information for entire job allocation (NodeList, ReqMem)
	allocation := lines[0]
	split := strings.Split(allocation, "|")
	metrics.JobName = split[0]
	metrics.User = split[1]
	metrics.Account = split[2]
	metrics.Partition = split[3]
	metrics.NodeList = split[4]
	metrics.State = split[6]
	cpuTimeStr := split[12]
	cpuTime := parseCpuTime(cpuTimeStr)
	if metrics.CPUTime < cpuTime {
		metrics.CPUTime = cpuTime
		metrics.CPUTimeStr = cpuTimeStr
	}
	totalCpuTimeStr := split[13]
	totalCpuTime := parseCpuTime(totalCpuTimeStr)
	if metrics.TotalCPU < totalCpuTime {
		metrics.TotalCPU = totalCpuTime
		metrics.TotalCPUStr = totalCpuTimeStr
	}
	userCpuTime := parseCpuTime(split[14])
	if metrics.UserCPU < userCpuTime {
		metrics.UserCPU = userCpuTime
	}
	systemCpuTime := parseCpuTime(split[15])
	if metrics.SystemCPU < systemCpuTime {
		metrics.SystemCPU = systemCpuTime
	}
	metrics.Nodes = len(strings.Split(split[4], ","))
	reqMem := strings.TrimSpace(split[16])
	if strings.HasSuffix(reqMem, "n") {
		metrics.ReqMem = uint64(metrics.Nodes) * parseByteSize(reqMem[:len(reqMem)-1])

	} else if strings.HasSuffix(reqMem, "c") {
		metrics.ReqMem = uint64(metrics.Ncpus) * parseByteSize(reqMem[:len(reqMem)-1])
	} else {
		metrics.ReqMem = parseByteSize(reqMem)
	}
	metrics.Submittime = split[7]
	metrics.Starttime = split[8]
	metrics.Endtime = split[9]
	metrics.WalltimeStr = split[10]
	metrics.Walltime = uint64(parseCpuTime(split[10]))
	metrics.RuntimeStr = split[11]
	metrics.Runtime = uint64(parseCpuTime(split[11]))

	l.Printf("Metrics: %#v", metrics)
	return &metrics, nil
}

func (m SacctMetrics) CalcUserComputePercentage() float64 {
	if m.TotalCPU != 0 {
		return (float64(m.UserCPU) / float64(m.TotalCPU)) * 100
	}
	return 0.0
}

func (m SacctMetrics) CalcSystemComputePercentage() float64 {
	if m.TotalCPU != 0 {
		return (float64(m.SystemCPU) / float64(m.TotalCPU)) * 100
	}
	return 0.0
}

func GetSacctMetrics(jobId string, paths map[string]string, l *log.Logger) (*SacctMetrics, error) {
	sacctMetrics, err := GetSacctData(jobId, paths, l)
	if err != nil {
		return nil, err
	}
	return ParseSacctMetrics(sacctMetrics, l)
}

func GetSstatMetrics(jobId string, paths map[string]string, l *log.Logger) (*SstatMetrics, error) {
	sstatMetrics, err := GetSacctData(jobId, paths, l)
	if err != nil {
		return nil, err
	}
	return ParseSstatMetrics(sstatMetrics)
}

func updateJobStatsWithLiveData(metrics *SacctMetrics, jobId string, paths map[string]string, l *log.Logger) {
	liveMetrics, err := GetSstatMetrics(jobId, paths, l)
	if err == nil {

		if liveMetrics.MaxRSS > 0 {
			metrics.MaxRSS = liveMetrics.MaxRSS
		}
		if liveMetrics.MaxDiskWrite > 0 {
			metrics.MaxDiskWrite = liveMetrics.MaxDiskWrite
		}
		if liveMetrics.MaxDiskRead > 0 {
			metrics.MaxDiskRead = liveMetrics.MaxDiskRead
		}
	}
}

// Execute the saccct command and return its output
func GetSacctData(jobId string, paths map[string]string, l *log.Logger) ([]byte, error) {
	formatLine := "JobName,User,Account,Partition,NodeList,ncpus,State,Submit,start,end,timelimit,elapsed,CPUTime,TotalCPU,UserCPU,SystemCPU,ReqMem,MaxRSS,MaxDiskWrite,MaxDiskRead,MaxRSSNode,MaxDiskWriteNode,MaxDiskReadNode,Comment"
	cmd := exec.Command(paths["sacct"], "-j", jobId, "-n", "-p", "--format", formatLine)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute sacct command: %w", err)
	}
	return output, nil
}

func GetSstatData(jobId string, paths map[string]string, l *log.Logger) ([]byte, error) {
	formatLine := "JobID,MaxRSS,MaxDiskWrite,MaxDiskRead,MaxRSSNode,MaxDiskWriteNode,MaxDiskReadNode"
	cmd := exec.Command(paths["sstat"], "-a", "-j", jobId, "-n", "-p", "--format", formatLine)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return nil, fmt.Errorf("failed to execute sstat command: %w", err)
	}
	return output, nil
}
