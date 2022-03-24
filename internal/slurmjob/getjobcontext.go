package slurmjob

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"
)

// TODO make it configureable or read it from sacctmgr

func calculateOptimalQOS(qosMap map[uint64]string, runtime uint64) string {
	keys := make([]uint64, len(qosMap))
	for k := range qosMap {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	for _, k := range keys {
		if runtime <= k {
			return qosMap[k]
		}
	}
	return "LONG"
}

// Populate JobContext.Hints with hints about the job (mis)usage and suggestions how to optimize it.
//	todo: add more logic once more stats are in
func (j *JobContext) GenerateHints(qosMap map[uint64]string) {
	if j.IsJobFinished() {
		// Check OUT_OF_MEMORY
		if j.SlurmEnvironment.SLURM_JOB_STATE == "OUT_OF_MEMORY" {
			j.Hints = append(j.Hints, "TIP: The job ran out of memory. Please re-submnit with increase memory requirements")
			return
		}

		if j.SlurmEnvironment.SLURM_JOB_STATE == "TIMEOUT" {
			j.Hints = append(j.Hints, "TIP: The job ran into a timeout. Please re-submnit with increase walltime requirements and potentially to a different QOS)")
			return
		}

		// Check memory consumption
		if j.JobStats.ReqMem/2 > j.JobStats.MaxRSS {
			j.Hints = append(j.Hints, "TIP: Please consider lowering the ammount of requested memory in the future, your job has consumed less then half of the requested memory.")
		}
		// check CPU time (16 cores requested only 1 used)
		if j.JobStats.CPUTime/2 > j.JobStats.TotalCPU {
			j.Hints = append(j.Hints, "TIP: Please consider lowering the amount of requested CPU cores in the future, your job has consumed less than half of requested CPU cores")
		}

		// Check if runtime is half of the requested runtime
		if j.JobStats.Walltime/2 > j.JobStats.Runtime {

			// Check if it was submitted without specifying a walltime (just against default maxwalltime of QOS)
			optimalQos := calculateOptimalQOS(qosMap, j.JobStats.Runtime)
			if qos, ok := qosMap[j.JobStats.Walltime]; ok {
				j.Hints = append(j.Hints, fmt.Sprintf("TIP: Your job was submitted to %s QOS and finished within half of the requested walltime. Consider submitting it to the %s QOS instead", qos, optimalQos))
				j.Hints = append(j.Hints, fmt.Sprintf("TIP: No --time specified: Using default %s QOS limit. Specify --time <walltime> to increase the chances that the scheduler will use this job for backfilling purposes! See https://docs.vbc.ac.at/link/21#bkmrk-scheduling-policy for more information.", qos))
			} else {

				j.Hints = append(j.Hints, fmt.Sprintf("TIP: Your job was submitted with a walltime of %s and finished in less half of the time, consider reducing the walltime and submitted to %s QOS", j.JobStats.WalltimeStr, optimalQos))
			}

		}
	}
}

// Get SLURM_* environment variables from the environment
//	todo: consider moving this to map and populate with keys matching "SLURM_"
func (j *JobContext) GetSlurmEnvVars() {

	j.SlurmEnvironment.SLURM_ARRAY_JOB_ID = os.Getenv("SLURM_ARRAY_JOB_ID")
	j.SlurmEnvironment.SLURM_ARRAY_TASK_COUNT = os.Getenv("SLURM_ARRAY_TASK_COUNT")
	j.SlurmEnvironment.SLURM_ARRAY_TASK_ID = os.Getenv("SLURM_ARRAY_TASK_ID")
	j.SlurmEnvironment.SLURM_ARRAY_TASK_MAX = os.Getenv("SLURM_ARRAY_TASK_MAX")
	j.SlurmEnvironment.SLURM_ARRAY_TASK_MIN = os.Getenv("SLURM_ARRAY_TASK_MIN")
	j.SlurmEnvironment.SLURM_ARRAY_TASK_STEP = os.Getenv("SLURM_ARRAY_TASK_STEP")
	j.SlurmEnvironment.SLURM_CLUSTER_NAME = os.Getenv("SLURM_CLUSTER_NAME")
	j.SlurmEnvironment.SLURM_JOB_ACCOUNT = os.Getenv("SLURM_JOB_ACCOUNT")
	j.SlurmEnvironment.SLURM_JOB_DERIVED_EC = os.Getenv("SLURM_JOB_DERIVED_EC")
	j.SlurmEnvironment.SLURM_JOB_EXIT_CODE = os.Getenv("SLURM_JOB_EXIT_CODE")
	j.SlurmEnvironment.SLURM_JOB_EXIT_CODE2 = os.Getenv("SLURM_JOB_EXIT_CODE2")
	j.SlurmEnvironment.SLURM_JOB_EXIT_CODE_MAX = os.Getenv("SLURM_JOB_EXIT_CODE_MAX")
	j.SlurmEnvironment.SLURM_JOB_EXIT_CODE_MIN = os.Getenv("SLURM_JOB_EXIT_CODE_MIN")
	j.SlurmEnvironment.SLURM_JOB_GID = os.Getenv("SLURM_JOB_GID")
	j.SlurmEnvironment.SLURM_JOB_GROUP = os.Getenv("SLURM_JOB_GROUP")
	j.SlurmEnvironment.SLURM_JOBID = os.Getenv("SLURM_JOBID")
	j.SlurmEnvironment.SLURM_JOB_ID = os.Getenv("SLURM_JOB_ID")
	j.SlurmEnvironment.SLURM_JOB_MAIL_TYPE = os.Getenv("SLURM_JOB_MAIL_TYPE")
	j.SlurmEnvironment.SLURM_JOB_NAME = os.Getenv("SLURM_JOB_NAME")
	j.SlurmEnvironment.SLURM_JOB_NODELIST = os.Getenv("SLURM_JOB_NODELIST")
	j.SlurmEnvironment.SLURM_JOB_PARTITION = os.Getenv("SLURM_JOB_PARTITION")
	j.SlurmEnvironment.SLURM_JOB_QUEUED_TIME = os.Getenv("SLURM_JOB_QUEUED_TIME")
	j.SlurmEnvironment.SLURM_JOB_RUN_TIME = os.Getenv("SLURM_JOB_RUN_TIME")
	j.SlurmEnvironment.SLURM_JOB_STATE = os.Getenv("SLURM_JOB_STATE")
	j.SlurmEnvironment.SLURM_JOB_STDIN = os.Getenv("SLURM_JOB_STDIN")
	j.SlurmEnvironment.SLURM_JOB_UID = os.Getenv("SLURM_JOB_UID")
	j.SlurmEnvironment.SLURM_JOB_USER = os.Getenv("SLURM_JOB_USER")
	j.SlurmEnvironment.SLURM_JOB_WORK_DIR = os.Getenv("SLURM_JOB_WORK_DIR")

}

func IsJobFinished(jobState string) bool {
	switch jobState {
	case
		"FAILED",
		"COMPLETED",
		"OUT_OF_MEMORY",
		"TIMEOUT":
		return true
	}
	return false
}

func (j *JobContext) IsJobFinished() bool {
	return IsJobFinished(j.SlurmEnvironment.SLURM_JOB_STATE)
}

// Get additional job statistics from external source (e.g. jobinfo or sacct)
func (j *JobContext) GetJobStats(log *log.Logger, subject string) {
	jobId := j.SlurmEnvironment.SLURM_JOBID
	if strings.Contains(subject, "Slurm Array Summary Job_id=") {
		j.MailSubject = fmt.Sprintf("Job Array Summary %s (%s-%s)", j.SlurmEnvironment.SLURM_ARRAY_JOB_ID, j.SlurmEnvironment.SLURM_ARRAY_TASK_MIN, j.SlurmEnvironment.SLURM_ARRAY_TASK_MAX)
	} else if strings.Contains(subject, "Slurm Array Task Job_id") {
		jobId = fmt.Sprintf("%s_%s", j.SlurmEnvironment.SLURM_ARRAY_JOB_ID, j.SlurmEnvironment.SLURM_ARRAY_TASK_ID)
		j.MailSubject = fmt.Sprintf("Job Array Task %s", jobId)

	} else {
		j.MailSubject = fmt.Sprintf("Job %s", jobId)

	}
	if j.SlurmEnvironment.SLURM_ARRAY_JOB_ID != "" {
		jobId = j.SlurmEnvironment.SLURM_ARRAY_JOB_ID
	}
	j.JobStats = *GetSacctMetrics(jobId, log)
	counter := 0
	for !IsJobFinished(j.JobStats.State) && j.JobStats.State != j.SlurmEnvironment.SLURM_JOB_STATE && counter < 5 {
		time.Sleep(2 * time.Second)
		j.JobStats = *GetSacctMetrics(jobId, log)
		counter += 1
	}
	if j.JobStats.State == "RUNNING" {
		updateJobStatsWithLiveData(&j.JobStats, jobId, log)
	}

	log.Printf("%#v", j.SlurmEnvironment)
}
