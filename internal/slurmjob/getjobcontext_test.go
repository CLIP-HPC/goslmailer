package slurmjob

import (
	"testing"
)

var qosMaxRuntimeMap = map[string]uint64{
	"RAPID":  3600,
	"SHORT":  28800,
	"MEDIUM": 172800,
	"LONG":   1209600,
}

func TestCalculateOptimalQos(t *testing.T) {
	qos := calculateOptimalQOS(qosMaxRuntimeMap, 600)
	if qos != "RAPID" {
		t.Errorf("Wrong QOS got: %s, want: RAPID", qos)
	}

	qos = calculateOptimalQOS(qosMaxRuntimeMap, 3800)
	if qos != "SHORT" {
		t.Errorf("Wrong QOS got: %s, want: SHORT", qos)
	}

	qos = calculateOptimalQOS(qosMaxRuntimeMap, 29000)
	if qos != "MEDIUM" {
		t.Errorf("Wrong QOS got: %s, want: MEDIUM", qos)
	}

	qos = calculateOptimalQOS(qosMaxRuntimeMap, 175000)
	if qos != "LONG" {
		t.Errorf("Wrong QOS got: %s, want: LONG", qos)
	}
}

func TestNoHintsWhenJobIsNotFinished(t *testing.T) {
	jobEnviron := SlurmEnvironment{SLURM_JOB_STATE: "RUNNING"}
	jobContext := JobContext{SlurmEnvironment: jobEnviron}
	jobContext.GenerateHints(qosMaxRuntimeMap)
	if len(jobContext.Hints) != 0 {
		t.Error("Running jobs should have now hints")
	}
}

func TestOOMHints(t *testing.T) {
	jobEnviron := SlurmEnvironment{SLURM_JOB_STATE: "OUT_OF_MEMORY"}
	jobContext := JobContext{SlurmEnvironment: jobEnviron}
	jobContext.GenerateHints(qosMaxRuntimeMap)
	if len(jobContext.Hints) != 1 {
		t.Errorf("Wrong number of hints. got: %d, want %d", len(jobContext.Hints), 1)
	}
}

func TestTimeOutHints(t *testing.T) {
	jobEnviron := SlurmEnvironment{SLURM_JOB_STATE: "TIMEOUT"}
	jobContext := JobContext{SlurmEnvironment: jobEnviron}
	jobContext.GenerateHints(qosMaxRuntimeMap)
	if len(jobContext.Hints) != 1 {
		t.Errorf("Wrong number of hints. got: %d, want %d", len(jobContext.Hints), 1)
	}
}

func TestRegularHints(t *testing.T) {
	jobEnviron := SlurmEnvironment{SLURM_JOB_STATE: "COMPLETED"}
	metrics := SacctMetrics{MaxRSS: 1610612736, ReqMem: 4294967296, Runtime: 11000, Walltime: 29000, CPUTime: 79488, TotalCPU: 30000}
	jobContext := JobContext{SlurmEnvironment: jobEnviron, JobStats: metrics}
	jobContext.GenerateHints(qosMaxRuntimeMap)
	if len(jobContext.Hints) != 3 {
		t.Errorf("Wrong number of hints. got: %d, want %d", len(jobContext.Hints), 3)
	}
}

func TestNoHints(t *testing.T) {
	jobEnviron := SlurmEnvironment{SLURM_JOB_STATE: "COMPLETED"}
	metrics := SacctMetrics{MaxRSS: 4294967296, ReqMem: 4294967296, Runtime: 22000, Walltime: 29000}
	jobContext := JobContext{SlurmEnvironment: jobEnviron, JobStats: metrics}
	jobContext.GenerateHints(qosMaxRuntimeMap)
	if len(jobContext.Hints) != 0 {
		t.Errorf("Wrong number of hints. got: %d, want %d", len(jobContext.Hints), 0)
	}
}

func TestSubjectParsing(t *testing.T) {
	// parse regular job BEGIN subject line
	subject := "Slurm Job_id=39766384 Name=rMPCD-PS:3.5_0 Began, Queued time 2-00:04:18"
	env, err := parseSubjectLine(subject)
	if err != nil {
		t.Fatalf("Error parsing subject: %s", err)
	}
	if env.SLURM_JOBID != "39766384" || env.SLURM_JOB_MAIL_TYPE != "Began" || env.SLURM_JOB_STATE != "RUNNING" || env.SLURM_JOB_NAME != "rMPCD-PS:3.5_0" {
		t.Errorf("jobid/MAIL_TYPE/JOB_STATE/JOB_NAME wrong. Got: %s/%s/%s/%s, want: 39766384/Began/RUNNING/rMPCD-PS:3.5_0", env.SLURM_JOBID, env.SLURM_JOB_MAIL_TYPE, env.SLURM_JOB_STATE, env.SLURM_JOB_NAME)
	}

	// parse task summary END job subject line
	subject = "Slurm Job_id=39891831 Name=L_R38_3 Ended, Run time 1-11:30:27, COMPLETED, ExitCode 0"
	env, err = parseSubjectLine(subject)
	if err != nil {
		t.Fatalf("Error parsing subject: %s", err)
	}
	if env.SLURM_JOBID != "39891831" || env.SLURM_JOB_MAIL_TYPE != "Ended" || env.SLURM_JOB_STATE != "COMPLETED" || env.SLURM_JOB_NAME != "L_R38_3" {
		t.Errorf("jobid/MAIL_TYPE/JOB_STATE/JOB_NAME wrong. Got: %s/%s/%s/%s, want: 39891831/Ended/COMPLETED/L_R38_3", env.SLURM_JOBID, env.SLURM_JOB_MAIL_TYPE, env.SLURM_JOB_STATE, env.SLURM_JOB_NAME)
	}

	// parse task summary BEGIN job subject line
	subject = "Slurm Array Summary Job_id=39860384_* (39860384) Name=count Began"
	env, err = parseSubjectLine(subject)
	if err != nil {
		t.Fatalf("Error parsing subject: %s", err)
	}
	if env.SLURM_JOBID != "39860384" || env.SLURM_JOB_MAIL_TYPE != "Began" || env.SLURM_JOB_STATE != "PENDING" || env.SLURM_JOB_NAME != "count" {
		t.Errorf("jobid/MAIL_TYPE wrong. Got: %s/%s/%s/%s, want: 39860384/Began/PENDING/count", env.SLURM_JOBID, env.SLURM_JOB_MAIL_TYPE, env.SLURM_JOB_STATE, env.SLURM_JOB_NAME)
	}

	// parse task summary END job subject line
	subject = "Slurm Array Summary Job_id=39973135_* (39973135) Name=2022_PLANTEEN_SCHIMPER_01_FC1_analysis.sbatch Ended, Mixed, MaxSignal [9]"
	env, err = parseSubjectLine(subject)
	if err != nil {
		t.Fatalf("Error parsing subject: %s", err)
	}
	if env.SLURM_JOBID != "39973135" || env.SLURM_JOB_MAIL_TYPE != "Ended" || env.SLURM_JOB_STATE != "Mixed" || env.SLURM_JOB_NAME != "2022_PLANTEEN_SCHIMPER_01_FC1_analysis.sbatch" {
		t.Errorf("jobid/MAIL_TYPE/JOB_STATE wrong. Got: %s/%s/%s/%s, want: 39973135/Ended/Mixed/2022_PLANTEEN_SCHIMPER_01_FC1_analysis.sbatch", env.SLURM_JOBID, env.SLURM_JOB_MAIL_TYPE, env.SLURM_JOB_STATE, env.SLURM_JOB_NAME)
	}

	// parse task array BEGIN job subject line
	subject = "Slurm Array Task Job_id=1052478_1 (1052479) Name=wrap Began, Queued time 00:00:01"
	env, err = parseSubjectLine(subject)
	if err != nil {
		t.Fatalf("Error parsing subject: %s", err)
	}
	if env.SLURM_JOBID != "1052479" || env.SLURM_ARRAY_JOB_ID != "1052478" || env.SLURM_JOB_MAIL_TYPE != "Began" || env.SLURM_JOB_STATE != "RUNNING" || env.SLURM_JOB_NAME != "wrap" {
		t.Errorf("jobid/jobarrayid/MAIL_TYPE wrong. Got: %s/%s/%s/%s/%s, want: 1052479/1052478/Began/RUNNING/wrap", env.SLURM_JOBID, env.SLURM_ARRAY_JOB_ID, env.SLURM_JOB_MAIL_TYPE, env.SLURM_JOB_STATE, env.SLURM_JOB_NAME)
	}

	// parse task array END job subject line
	subject = "Slurm Array Task Job_id=1052478_1 (1052479) Name=wrap Ended, Run time 00:00:08, COMPLETED, ExitCode 0"
	env, err = parseSubjectLine(subject)
	if err != nil {
		t.Fatalf("Error parsing subject: %s", err)
	}
	if env.SLURM_JOBID != "1052479" || env.SLURM_ARRAY_JOB_ID != "1052478" || env.SLURM_JOB_MAIL_TYPE != "Ended" || env.SLURM_JOB_STATE != "COMPLETED" || env.SLURM_JOB_NAME != "wrap" {
		t.Errorf("jobid/jobarrayid/MAIL_TYPE/JOB_STATE wrong. Got: %s/%s/%s/%s/%s, want: 1052479/1052478/Ended/COMPLETED/wrap", env.SLURM_JOBID, env.SLURM_ARRAY_JOB_ID, env.SLURM_JOB_MAIL_TYPE, env.SLURM_JOB_STATE, env.SLURM_JOB_NAME)
	}

	// parse error message when wrong subject line
	subject = "slurm job x"
	env, err = parseSubjectLine(subject)
	if err == nil {
		t.Fatalf("No error thrown for wrong subject")
	}
}
