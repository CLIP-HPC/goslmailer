package slurmjob

type SlurmEnvironment struct {
	SLURM_ARRAY_JOB_ID      string
	SLURM_ARRAY_TASK_COUNT  string
	SLURM_ARRAY_TASK_ID     string
	SLURM_ARRAY_TASK_MAX    string
	SLURM_ARRAY_TASK_MIN    string
	SLURM_ARRAY_TASK_STEP   string
	SLURM_CLUSTER_NAME      string
	SLURM_JOB_ACCOUNT       string
	SLURM_JOB_DERIVED_EC    string
	SLURM_JOB_EXIT_CODE     string
	SLURM_JOB_EXIT_CODE2    string
	SLURM_JOB_EXIT_CODE_MAX string
	SLURM_JOB_EXIT_CODE_MIN string
	SLURM_JOB_GID           string
	SLURM_JOB_GROUP         string
	SLURM_JOBID             string
	SLURM_JOB_ID            string
	SLURM_JOB_MAIL_TYPE     string
	SLURM_JOB_NAME          string
	SLURM_JOB_NODELIST      string
	SLURM_JOB_PARTITION     string
	SLURM_JOB_QUEUED_TIME   string
	SLURM_JOB_RUN_TIME      string
	SLURM_JOB_STATE         string
	SLURM_JOB_STDIN         string
	SLURM_JOB_UID           string
	SLURM_JOB_USER          string
	SLURM_JOB_WORK_DIR      string
}

type JobContext struct {
	SlurmEnvironment
        JobStats SacctMetrics
        Hints    []string
}
