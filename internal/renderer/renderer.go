package renderer

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"os"
	texttemplate "text/template"
	"time"

	"github.com/CLIP-HPC/goslmailer/internal/slurmjob"
	"github.com/dustin/go-humanize"
)

// RenderTemplate renders the template file 'tfile' into 'buf' Buffer, using 'format' go package ('HTML' for html/template, 'text' for text/template).
// 'j slurmjob.JobContext' and 'userid string' are wrapped in a structure to be used as template data.
func RenderTemplate(tfile, format string, j *slurmjob.JobContext, userid string, buf *bytes.Buffer) error {

	var x = struct {
		Job     slurmjob.JobContext
		UserID  string
		Created string
	}{
		*j,
		userid,
		fmt.Sprint(time.Now().Format("Mon, 2 Jan 2006 15:04:05 MST")),
	}

	// get template file
	f, err := os.ReadFile(tfile)
	if err != nil {
		return err
	}

	// depending on `format` (from connector map in conf file), render template
	if format == "HTML" {
		var funcMap = htmltemplate.FuncMap{
			"humanBytes": humanize.Bytes,
		}
		t := htmltemplate.Must(htmltemplate.New(tfile).Funcs(funcMap).Parse(string(f)))
		err = t.Execute(buf, x)
	} else {
		var funcMap = texttemplate.FuncMap{
			"humanBytes": humanize.Bytes,
		}
		t := texttemplate.Must(texttemplate.New(tfile).Funcs(funcMap).Parse(string(f)))
		err = t.Execute(buf, x)
	}
	return err
}
