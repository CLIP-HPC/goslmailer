package renderer

import (
	"bytes"
	"fmt"
	htmltemplate "html/template"
	"os"
	texttemplate "text/template"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/pja237/goslmailer/internal/slurmjob"
)

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
