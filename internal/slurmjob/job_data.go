package slurmjob

type JobContext struct {
	SlurmEnvironment
	JobStats           SacctMetrics
	Hints              []string
	MailSubject        string
	PrunedMessageCount uint32
}

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

type SacctMetrics struct {
	JobName      string
	User         string
	Account      string
	Partition    string
	State        string
	Ncpus        int64
	Nodes        int
	NodeList     string
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
