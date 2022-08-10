# Templating guide

> **Info**
> In SLURM < 21.08.x, only a subset of job related information are available as SLURM environment variables in the `adaptive_card_template.json` and `telegramTemplate.html` templates. Instead of the SLURM environment variables (i.e `Job.SlurmEnvironment.SLURM_JOB_USER`) the variables from `SacctMetrics` can be used (i.e. `.Job.JobStats.User`) instead. See the [adaptive_card_template.json](test_e2e/cases/test_05/conf/adaptive_card_template.json) in the `test_e2e/cases/test_05` test case as an example.

Goslmailer uses golang [text/template](https://pkg.go.dev/text/template) and [html/template](https://pkg.go.dev/html/template) libraries.

The connectors call `renderer.RenderTemplate` function.

Data structure you can reference in the template can be found in:

* [job_data.go](../internal/slurmjob/job_data.go)
* [rendererer.go](../internal/renderer/renderer.go)

Example:

* `{{ .Job.SlurmEnvironment.SLURM_JOB_MAIL_TYPE }}`
* `{{ .Job.JobStats.MaxRSS | humanBytes }}`

* [Example telegram html template](./telegramTemplate.html)
* [More template examples](./templates/)

Structures:

```
struct {
  Job     slurmjob.JobContext
  UserID  string
  Created string
}

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
```
