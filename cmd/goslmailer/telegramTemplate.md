{{ .Job.MailSubject }} {{ .Job.SlurmEnvironment.SLURM_JOB_MAIL_TYPE }}
`----------------------------------------`
{{ if ne .Job.PrunedMessageCount 0 }}
*WARNING: Rate limiting triggered. {{ .Job.PrunedMessageCount }} additonal notificiations have been suppressed*
`----------------------------------------`
{{ end }}
```
Job Name          : {{ .Job.SlurmEnvironment.SLURM_JOB_NAME }}
Job ID            : {{ .Job.SlurmEnvironment.SLURM_JOB_ID }}
User              : {{ .Job.SlurmEnvironment.SLURM_JOB_USER }}
Partition         : {{ .Job.SlurmEnvironment.SLURM_JOB_PARTITION }}
Nodes Used        : {{ .Job.SlurmEnvironment.SLURM_JOB_NODELIST }}
Cores             : {{ .Job.JobStats.Ncpus }}
Job state         : {{ .Job.SlurmEnvironment.SLURM_JOB_STATE }}
{{- if ne .Job.SlurmEnvironment.SLURM_JOB_STATE "RUNNING" }}
Exit Code         : {{ .Job.SlurmEnvironment.SLURM_JOB_EXIT_CODE_MAX }}
{{- end }}
Submit            : {{ .Job.JobStats.Submittime }}
Start             : {{ .Job.JobStats.Starttime }}
{{- if ne .Job.SlurmEnvironment.SLURM_JOB_STATE "RUNNING" }}
End               : {{ .Job.JobStats.Endtime }}
{{- end }}
Reserved Walltime : {{ .Job.JobStats.WalltimeStr }}
{{- if ne .Job.SlurmEnvironment.SLURM_JOB_MAIL_TYPE "Began" }}
Used Walltime     : {{ .Job.SlurmEnvironment.SLURM_JOB_RUN_TIME }}
{{- if ne .Job.SlurmEnvironment.SLURM_JOB_STATE "RUNNING" }}
Used CPU time     : {{ .Job.JobStats.TotalCPUStr }}
% User (Computation) : {{ printf "%5.2f%%" .Job.JobStats.CalcUserComputePercentage }}
% System (I/O)    : {{ printf "%5.2f%%" .Job.JobStats.CalcSystemComputePercentage }}
{{- end }}
{{- end }}
Memory Requested  : {{ .Job.JobStats.ReqMem | humanBytes }}
{{- if ne .Job.SlurmEnvironment.SLURM_JOB_MAIL_TYPE "Began" }}
Max Memory Used   : {{ .Job.JobStats.MaxRSS | humanBytes }}
Max Disk Write    : {{ .Job.JobStats.MaxDiskWrite | humanBytes }}
Max Disk Read     : {{ .Job.JobStats.MaxDiskRead | humanBytes }}
{{- end }}
```
`----------------------------------------`
```
{{- range .Job.Hints }}
{{ . }}
{{- end }}
```
`----------------------------------------`
