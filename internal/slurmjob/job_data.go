package slurmjob

type SlurmEnvironment struct {
	SLURM_JOB_NAME          string
	SLURM_JOB_GROUP         string
	SLURM_JOB_STATE         string
	SLURM_JOB_WORK_DIR      string
	SLURM_JOB_MAIL_TYPE     string
	SLURM_JOBID             string
	SLURM_JOB_RUN_TIME      string
	SLURM_JOB_EXIT_CODE2    string
	SLURM_JOB_DERIVED_EC    string
	SLURM_JOB_ID            string
	SLURM_JOB_USER          string
	SLURM_JOB_EXIT_CODE     string
	SLURM_JOB_UID           string
	SLURM_JOB_NODELIST      string
	SLURM_JOB_EXIT_CODE_MAX string
	SLURM_JOB_GID           string
	SLURM_CLUSTER_NAME      string
	SLURM_JOB_PARTITION     string
	SLURM_JOB_ACCOUNT       string
}

type JobStats struct {
	MemReq, MemUsed int64
}

type JobContext struct {
	SlurmEnvironment
	JobStats
	Hints []string
}
