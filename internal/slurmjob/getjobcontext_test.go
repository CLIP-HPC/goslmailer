package slurmjob

import (
        "testing"
)

var qosMaxRuntimeMap = map[uint64]string{
        3600:    "RAPID",
        28800:   "SHORT",
        172800:  "MEDIUM",
        1209600: "LONG",
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
