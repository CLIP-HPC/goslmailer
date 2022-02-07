package slurmjob

import "os"

// Populate JobContext.Hints with hints about the job (mis)usage and suggestions how to optimize it.
//	todo: add more logic once more stats are in
func (j *JobContext) GenerateHints() {
	if j.JobStats.MemReq/2 > j.JobStats.MemUsed {
		j.Hints = append(j.Hints, "TIP: Please consider lowering the ammount of requested memory in the future, your job has consumed less then half of requested.")
	}
	j.Hints = append(j.Hints, "TIP: Your job was submitted to LONG queue and finished in less then an hour, consider submitting to RAPID or SHORT queues.")
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

// Get additional job statistics from external source (e.g. jobinfo or sacct)
func (j *JobContext) GetJobStats(s SlurmEnvironment) {
	j.JobStats.MemReq = 4096
	j.JobStats.MemUsed = 1024
	// FUTURE: employ jobinfo to do the job or call sacct and get the data we need ourself
	//
	//out, err := exec.Command("./jobinfo", string(s.SLURM_JOBID)).Output()
	//if err != nil {
	//	fmt.Println("ERROR Executing jobinfo. Abort!")
	//	os.Exit(1)
	//}
	////fmt.Println(string(out))
	////fmt.Printf("%#v\n", strings.Split(string(out), "\n"))
	//for _, l := range strings.Split(string(out), "\n") {
	//	//fmt.Printf("PRE SPLIT %#v\n", l)
	//	v := strings.Split(l, ":")
	//	fmt.Printf("POST SPLIT %#v\n", v)
	//	//fmt.Printf("ACCESS %#v : %#v\n", strings.Trim(v[0], " "), strings.Trim(v[1], " "))
	//}
}
