package slurmjob

import (
	"io/ioutil"
	"os"
	"testing"
)

func TestParseSacctMetrics(t *testing.T) {
	// Read the input data from a file
	file, err := os.Open("../../test_data/sacct.txt")
	if err != nil {
		t.Fatalf("Can not open test data: %v", err)
	}
	data, _ := ioutil.ReadAll(file)
	metrics, _ := ParseSacctMetrics(data)
	t.Logf("%+v", metrics)

	if metrics.JobName != "JobName" {
		t.Errorf("JobName is incorrect. got: %s, want: %s", metrics.JobName, "JobName")
	}

	if metrics.User != "username" {
		t.Errorf("User is incorrect. got: %s, want: %s", metrics.User, "username")
	}
	if metrics.Account != "account" {
		t.Errorf("Account is incorrect. got: %s, want: %s", metrics.Account, "account")
	}
	if metrics.Partition != "c" {
		t.Errorf("Partition is incorrect. got: %s, want: %s", metrics.Partition, "c")
	}
	if metrics.NodeList != "clip-c2-10" {
		t.Errorf("Partition is incorrect. got: %s, want: %s", metrics.NodeList, "clip-c2-10")
	}
	if metrics.State != "COMPLETED" {
		t.Errorf("State is incorrect. got: %s, want: %s", metrics.State, "COMPLETED")
	}
	if metrics.Nodes != 1 {
		t.Errorf("Nodes is incorrect. got: %d, want: %d", metrics.Nodes, 1)
	}
	if metrics.Ncpus != 4 {
		t.Errorf("Ncpus is incorrect. got: %d, want: %d", metrics.Ncpus, 4)
	}
	if metrics.Submittime != "2022-02-16T20:40:15" {
		t.Errorf("Submittime is incorrect. got: %s, want: %s", metrics.Submittime, "2022-02-16T20:40:15")
	}
	if metrics.Starttime != "2022-02-16T20:40:15" {
		t.Errorf("Starttime is incorrect. got: %s, want: %s", metrics.Starttime, "2022-02-16T20:40:15")
	}
	if metrics.Endtime != "2022-02-17T01:11:04" {
		t.Errorf("Endtime is incorrect. got: %s, want: %s", metrics.Endtime, "2022-02-17T01:11:04")
	}
	if metrics.CPUTimeStr != "18:03:16" {
		t.Errorf("CPUTimeStr is incorrect. got: %s, want: %s", metrics.CPUTimeStr, "18:03:16")
	}
	if metrics.CPUTime != 64996 {
		t.Errorf("CPUTime is incorrect. got: %f, want: %f", metrics.CPUTime, 64996.0)
	}
	if metrics.TotalCPUStr != "01:57.511" {
		t.Errorf("TotalCPUStr is incorrect. got: %s, want: %s", metrics.TotalCPUStr, "01:57.511")
	}
	if metrics.TotalCPU != 117.511 {
		t.Errorf("TotalCPU is incorrect. got: %f, want: %f", metrics.TotalCPU, 117.511)
	}
	if metrics.UserCPU != 102.011 {
		t.Errorf("UserCPU is incorrect. got: %f, want: %f", metrics.UserCPU, 102.011)
	}
	if metrics.SystemCPU != 15.5 {
		t.Errorf("SystemCPU is incorrect. got: %f, want: %f", metrics.SystemCPU, 15.5)
	}
	if metrics.ReqMem != 34359738368 {
		t.Errorf("ReqMem is incorrect. got: %d, want: %d", metrics.ReqMem, 34359738368)
	}
	if metrics.MaxRSS != 1133199360 {
		t.Errorf("MaxRSS is incorrect. got: %d, want: %d", metrics.MaxRSS, 1133199360)
	}
	if metrics.WalltimeStr != "08:00:00" {
		t.Errorf("WalltimeStr is incorrect. got: %s, want: %s", metrics.WalltimeStr, "08:00:00")
	}
	if metrics.Walltime != 28800 {
		t.Errorf("Walltime is incorrect. got: %d, want: %d", metrics.Walltime, 28800)
	}
	if metrics.RuntimeStr != "04:30:49" {
		t.Errorf("RuntimeStr is incorrect. got: %s, want: %s", metrics.RuntimeStr, "04:30:49")
	}
	if metrics.Runtime != 16249 {
		t.Errorf("Runtime is incorrect. got: %d, want: %d", metrics.Runtime, 16249)
	}
	if metrics.MaxRSS != 1133199360 {
		t.Errorf("MaxRSS is incorrect. got: %d, want: %d", metrics.MaxRSS, 1133199360)
	}
	if metrics.MaxDiskWrite != 10485 {
		t.Errorf("MaxDiskWrite is incorrect. got: %d, want: %d", metrics.MaxDiskWrite, 10485)
	}
	if metrics.MaxDiskRead != 136314 {
		t.Errorf("MaxDiskRead is incorrect. got: %d, want: %d", metrics.MaxDiskRead, 136314)
	}
}

func TestParseSstatMetrics(t *testing.T) {
	// Read the input data from a file
	file, err := os.Open("../../test_data/sstat.txt")
	if err != nil {
		t.Fatalf("Can not open test data: %v", err)
	}
	data, _ := ioutil.ReadAll(file)
	metrics, _ := ParseSstatMetrics(data)
	t.Logf("%+v", metrics)
	if metrics.MaxRSS != 1850245120 {
		t.Errorf("MaxRSS is incorrect. got: %d, want: %d", metrics.MaxRSS, 1850245120)
	}
	if metrics.MaxDiskWrite != 70 {
		t.Errorf("MaxDiskWrite is incorrect. got: %d, want: %d", metrics.MaxDiskWrite, 70)
	}
	if metrics.MaxDiskRead != 205384 {
		t.Errorf("MaxDiskRead is incorrect. got: %d, want: %d", metrics.MaxDiskRead, 205384)
	}
}

func TestParseSacctMetricsEmptyInput(t *testing.T) {
	// Read the input data from a file
	metrics, _ := ParseSacctMetrics([]byte(""))
	var emptyMetrics SacctMetrics
	t.Logf("%+v", metrics)

	if *metrics != emptyMetrics {
		t.Error("Empty input should return empty metrics")
	}

}
