package slurmjob

import (
        "log"
        "math"
        "os/exec"
        "regexp"
        "strconv"
        "strings"
)

type SacctMetrics struct {
        State        string
        Ncpus        int64
        Nodes        int
        Submittime   string
        Starttime    string
        Endtime      string
        CPUTimeStr   string
        CPUTime      float64
        TotalCPU     float64
        TotalCPUStr  string
        UserCPU      float64
        SystemCPU    float64
        ReqMem       uint64
        MaxRSS       uint64
        Walltime     uint64
        WalltimeStr  string
        Runtime      uint64
        RuntimeStr   string
        MaxDiskWrite uint64
        MaxDiskRead  uint64
}

type SstatMetrics struct {
        MaxRSS       uint64
        MaxDiskWrite uint64
        MaxDiskRead  uint64
}

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

func ParseSstatMetrics(input []byte) *SstatMetrics {
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
        return &metrics
}

func ParseSacctMetrics(input []byte) *SacctMetrics {
        var metrics SacctMetrics
        if len(input) == 0 {
                return &metrics
        }
        lines := strings.Split(string(input), "\n")
        for _, line := range lines {
                if line == "" {
                        continue
                }
                split := strings.Split(line, "|")
                ncpus, _ := strconv.ParseInt(strings.TrimSpace(split[4]), 10, 16)
                if metrics.Ncpus < ncpus {
                        metrics.Ncpus = ncpus
                }
                if split[16] != "" {
                        maxRSS := parseByteSize(split[16])
                        if metrics.MaxRSS < maxRSS {
                                metrics.MaxRSS = maxRSS
                        }
                }
                if split[17] != "" {
                        maxDiskWrite := parseByteSize(split[17])
                        if metrics.MaxDiskWrite < maxDiskWrite {
                                metrics.MaxDiskWrite = maxDiskWrite
                        }
                }
                if split[18] != "" {
                        maxDiskRead := parseByteSize(split[18])
                        if metrics.MaxDiskRead < maxDiskRead {
                                metrics.MaxDiskRead = maxDiskRead
                        }
                }

        }
        // retrieve information for entire job allocation (NodeList, ReqMem)
        allocation := lines[0]
        split := strings.Split(allocation, "|")
        metrics.State = split[5]
        cpuTimeStr := split[11]
        cpuTime := parseCpuTime(cpuTimeStr)
        if metrics.CPUTime < cpuTime {
                metrics.CPUTime = cpuTime
                metrics.CPUTimeStr = cpuTimeStr
        }
        totalCpuTimeStr := split[12]
        totalCpuTime := parseCpuTime(totalCpuTimeStr)
        if metrics.TotalCPU < totalCpuTime {
                metrics.TotalCPU = totalCpuTime
                metrics.TotalCPUStr = totalCpuTimeStr
        }
        userCpuTime := parseCpuTime(split[13])
        if metrics.UserCPU < userCpuTime {
                metrics.UserCPU = userCpuTime
        }
        systemCpuTime := parseCpuTime(split[14])
        if metrics.SystemCPU < systemCpuTime {
                metrics.SystemCPU = systemCpuTime
        }
        metrics.Nodes = len(strings.Split(split[3], ","))
        reqMem := strings.TrimSpace(split[15])
        if strings.HasSuffix(reqMem, "n") {
                metrics.ReqMem = uint64(metrics.Nodes) * parseByteSize(reqMem[:len(reqMem)-1])

        } else if strings.HasSuffix(reqMem, "c") {
                metrics.ReqMem = uint64(metrics.Ncpus) * parseByteSize(reqMem[:len(reqMem)-1])
        } else {
                metrics.ReqMem = parseByteSize(reqMem)
        }
        metrics.Submittime = split[6]
        metrics.Starttime = split[7]
        metrics.Endtime = split[8]
        metrics.WalltimeStr = split[9]
        metrics.Walltime = uint64(parseCpuTime(split[9]))
        metrics.RuntimeStr = split[10]
        metrics.Runtime = uint64(parseCpuTime(split[10]))

        log.Printf("Metrics: %#v", metrics)
        return &metrics
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

func GetSacctMetrics(jobId string, log *log.Logger, paths map[string]string) *SacctMetrics {
        return ParseSacctMetrics(GetSacctData(jobId, log, paths))
}

func GetSstatMetrics(jobId string, log *log.Logger, paths map[string]string) *SstatMetrics {
        return ParseSstatMetrics(GetSstatData(jobId, log, paths))
}

func updateJobStatsWithLiveData(metrics *SacctMetrics, jobId string, log *log.Logger, paths map[string]string) {
        liveMetrics := GetSstatMetrics(jobId, log, paths)
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

// Execute the saccct command and return its output
func GetSacctData(jobId string, log *log.Logger, paths map[string]string) []byte {
        formatLine := "JobName,User,Partition,NodeList,ncpus,State,Submit,start,end,timelimit,elapsed,CPUTime,TotalCPU,UserCPU,SystemCPU,ReqMem,MaxRSS,MaxDiskWrite,MaxDiskRead,MaxRSSNode,MaxDiskWriteNode,MaxDiskReadNode,Comment"
        cmd := exec.Command(paths["sacct"], "-j", jobId, "-n", "-p", "--format", formatLine)
        output, err := cmd.CombinedOutput()
        if err != nil {
                log.Fatal(output)
        }
        return output
}

func GetSstatData(jobId string, log *log.Logger, paths map[string]string) []byte {
        formatLine := "JobID,MaxRSS,MaxDiskWrite,MaxDiskRead,MaxRSSNode,MaxDiskWriteNode,MaxDiskReadNode"
        cmd := exec.Command(paths["sstat"], "-a", "-j", jobId, "-n", "-p", "--format", formatLine)
        output, err := cmd.CombinedOutput()
        if err != nil {
                log.Fatal(output)
        }
        return output
}
