{{ .Job.MailSubject }} {{ .Job.SlurmEnvironment.SLURM_JOB_MAIL_TYPE }}
`----------------------------------------`
{{ if ne .Job.PrunedMessageCount 0 }}
WARNING: Rate limiting triggered. {{ .Job.PrunedMessageCount }} additonal notificiations have been suppressed
`----------------------------------------`
{{ end }}

```
Job Name : {{ .Job.SlurmEnvironment.SLURM_JOB_NAME }}
Job ID   : {{ .Job.SlurmEnvironment.SLURM_JOB_ID }}
User     : {{ .Job.SlurmEnvironment.SLURM_JOB_USER }}
```
`----------------------------------------`
