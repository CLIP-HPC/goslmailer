package slurmjob

import (
	"errors"
	"fmt"
	"log"
	"os"
	"regexp"
	"sort"
	"strings"
	"time"
)

type qosMapQL map[string]uint64 // qosMap[QoS]Limit
type qosMapLQ map[uint64]string // qosMap[Limit]QoS

// makeQoSReversed reverses qosMap from config map[string]uint64 to map[uint64]string format.
func makeQoSReversed(ql qosMapQL) qosMapLQ {

	lq := make(map[uint64]string)

	for k, v := range ql {
		lq[v] = k
	}

	return lq
}

// TODO make it configureable or read it from sacctmgr

func calculateOptimalQOS(qosMap qosMapQL, runtime uint64) string {

	// make a slice of structs used to sort the map
	keys := make([]struct {
		qos string // qos name
		lim uint64 // time limit
	}, len(qosMap))

	// put everything inside
	for k, v := range qosMap {
		keys = append(keys, struct {
			qos string
			lim uint64
		}{qos: k, lim: v})
	}

	// sort it
	sort.Slice(keys, func(i, j int) bool { return keys[i].lim < keys[j].lim })

	// find fitting qos for the job
	for _, k := range keys {
		if runtime <= k.lim {
			return k.qos
		}
	}

	return "LONG"
	// or
	//return keys[len(keys)-1].qos
}

// Populate JobContext.Hints with hints about the job (mis)usage and suggestions how to optimize it.
//	todo: add more logic once more stats are in
func (j *JobContext) GenerateHints(qosMap qosMapQL) {
	if j.IsJobFinished() {
		// Check OUT_OF_MEMORY
		if j.SlurmEnvironment.SLURM_JOB_STATE == "OUT_OF_MEMORY" {
			j.Hints = append(j.Hints, "TIP: The job ran out of memory. Please re-submit with increased memory requirements")
			return
		}

		if j.SlurmEnvironment.SLURM_JOB_STATE == "TIMEOUT" {
			j.Hints = append(j.Hints, "TIP: The job ran into a timeout. Please re-submit with increased walltime requirements and potentially to a different QOS)")
			return
		}

		// Check memory consumption
		if j.JobStats.ReqMem/2 > j.JobStats.MaxRSS {
			j.Hints = append(j.Hints, "TIP: Please consider lowering the amount of requested memory in the future, your job has consumed less than half of the requested memory.")
		}
		// check CPU time (16 cores requested only 1 used)
		if j.JobStats.CPUTime/2 > j.JobStats.TotalCPU {
			j.Hints = append(j.Hints, "TIP: Please consider lowering the amount of requested CPU cores in the future, your job has consumed less than half of the requested CPU cores")
		}

		// Check if runtime is half of the requested runtime
		if j.JobStats.Walltime/2 > j.JobStats.Runtime {

			// Check if it was submitted without specifying a walltime (just against default maxwalltime of QOS)
			optimalQos := calculateOptimalQOS(qosMap, j.JobStats.Runtime)

			// reverse the map to LimitQos format map[uint64]string to be used onwards without too many modifications to code...
			lqMap := makeQoSReversed(qosMap)
			//if qos, ok := qosMap[j.JobStats.Walltime]; ok { // original
			if qos, ok := lqMap[j.JobStats.Walltime]; ok { // new
				if qos != optimalQos {
					j.Hints = append(j.Hints, fmt.Sprintf("TIP: Your job was submitted to %s QOS and finished within half of the requested walltime. Consider submitting it to the %s QOS instead", qos, optimalQos))
				} else {
					j.Hints = append(j.Hints, fmt.Sprintf("TIP: Your job was submitted to %s QOS and finished within half of the requested walltime. Consider reducing the walltime for backfilling purposes", qos))
				}
				j.Hints = append(j.Hints, fmt.Sprintf("TIP: No --time specified: Using default %s QOS limit. Specify --time to increase the chances that the scheduler will use this job for backfilling purposes", qos))
			} else {

				j.Hints = append(j.Hints, fmt.Sprintf("TIP: Your job was submitted with a walltime of %s and finished in less half of the time, consider reducing the walltime and submit it to %s QOS", j.JobStats.WalltimeStr, optimalQos))
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
		"TIMEOUT",
		"Mixed":
		return true
	}
	return false
}

func (j *JobContext) IsJobFinished() bool {
	return IsJobFinished(j.SlurmEnvironment.SLURM_JOB_STATE)
}

// Parse a subject line string and return a partial filled SlurmEnvirinment struct
// throw error if parsing failes
func parseSubjectLine(subject string) (*SlurmEnvironment, error) {
	rJob, _ := regexp.Compile(`^Slurm Job_id=(?P<JobId>\d+) Name=(?P<JobName>.*) (?P<MailType>\w+), \w+ time .+?(?:, (?P<JobState>\w+), ExitCode (?P<ExitCode>\d))?$`)
	asJob, _ := regexp.Compile(`^Slurm Array Summary Job_id=.+ \((\d+)\) Name=(?P<JobName>.*) (?P<MailType>\w+)(?:, (?P<JobState>\w+), \w+ \[(?P<ExitCode>.+)\])?$`)
	aJob, _ := regexp.Compile(`^Slurm Array Task Job_id=(?P<JobArrayId>\d+)_(?P<JobArrayIndex>\d+) \((?P<JobId>\d+)\) Name=(?P<JobName>.*) (?P<MailType>\w+), \w+ time .+?(?:, (?P<JobState>\w+), ExitCode (?P<ExitCode>\d))?$`)

	env := new(SlurmEnvironment)
	var jobId string
	var jobState string
	var mailType string
	var jobName string
	if strings.Contains(subject, "Slurm Array Summary Job_id=") {
		matches := asJob.FindStringSubmatch(subject)
		if matches == nil {
			return nil, errors.New(("Invalid subject line: " + subject))
		}
		jobId = matches[1]
		jobName = matches[2]
		mailType = matches[3]
		jobState = matches[4]
		if jobState == "" {
			jobState = "PENDING"
		}
	} else if strings.Contains(subject, "Slurm Array Task Job_id") {
		matches := aJob.FindStringSubmatch(subject)
		if matches == nil {
			return nil, errors.New(("Invalid subject line: " + subject))
		}
		env.SLURM_ARRAY_JOB_ID = matches[1]
		env.SLURM_ARRAY_TASK_ID = matches[2]
		jobId = matches[3]
		jobName = matches[4]
		mailType = matches[5]
		jobState = matches[6]
		if jobState == "" {
			jobState = "RUNNING"
		}
	} else {
		matches := rJob.FindStringSubmatch(subject)
		if matches == nil {
			return nil, errors.New(("Invalid subject line: " + subject))
		}
		jobId = matches[1]
		jobName = matches[2]
		mailType = matches[3]
		jobState = matches[4]
		if jobState == "" {
			jobState = "RUNNING"
		}

	}
	env.SLURM_JOBID = jobId
	env.SLURM_JOB_ID = jobId
	env.SLURM_JOB_MAIL_TYPE = mailType
	env.SLURM_JOB_STATE = jobState
	env.SLURM_JOB_NAME = jobName
	return env, nil
}

func (j *JobContext) UpdateEnvVarsFromMailSubject(subject string) error {
	env, err := parseSubjectLine(subject)
	if err != nil {
		return err
	}
	j.SlurmEnvironment = *env
	return nil
}

// Get additional job statistics from external source (e.g. jobinfo or sacct)
func (j *JobContext) GetJobStats(subject string, paths map[string]string, l *log.Logger) error {
	l.Print("Start retrieving job stats")
	l.Printf("%#v", j.SlurmEnvironment)

	// SLURM < 21.08.x don't have any SLURM envs set, we need to parse the mail subject line, retrieve the jobid and all other information from sacct
	if j.SlurmEnvironment.SLURM_JOBID == "" {
		err := j.UpdateEnvVarsFromMailSubject(subject)
		if err != nil {
			return err
		}
	}
	jobId := j.SlurmEnvironment.SLURM_JOBID
	if strings.Contains(subject, "Slurm Array Summary Job_id=") {
		j.MailSubject = fmt.Sprintf("Job Array Summary %s_*", j.SlurmEnvironment.SLURM_ARRAY_JOB_ID)
	} else if strings.Contains(subject, "Slurm Array Task Job_id") {
		j.MailSubject = fmt.Sprintf("Job Array Task %s", jobId)

	} else {
		j.MailSubject = fmt.Sprintf("Job %s", jobId)
	}
	if j.SlurmEnvironment.SLURM_ARRAY_JOB_ID != "" {
		jobId = j.SlurmEnvironment.SLURM_ARRAY_JOB_ID
	}
	l.Printf("Fetch job info %s", jobId)
	jobStats, err := GetSacctMetrics(jobId, paths, l)
	if err != nil {
		return err
	}
	j.JobStats = *jobStats
	counter := 0
	for !IsJobFinished(j.JobStats.State) && j.JobStats.State != j.SlurmEnvironment.SLURM_JOB_STATE && counter < 5 {
		time.Sleep(2 * time.Second)
		jobStats, err = GetSacctMetrics(jobId, paths, l)
		if err != nil {
			return fmt.Errorf("failed to get job stats: %w", err)
		}
		j.JobStats = *jobStats
		counter += 1
	}
	if j.JobStats.State == "RUNNING" {
		l.Print("Update job with live stats")
		updateJobStatsWithLiveData(&j.JobStats, jobId, paths, l)
	}
	l.Printf("Finished retrieving job stats")
	return nil
}
