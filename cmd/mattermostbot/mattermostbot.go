package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/CLIP-HPC/goslmailer/internal/cmdline"
	"github.com/CLIP-HPC/goslmailer/internal/config"
	"github.com/CLIP-HPC/goslmailer/internal/logger"
	"github.com/CLIP-HPC/goslmailer/internal/version"
	"github.com/mattermost/mattermost-server/v5/model"
)

const app = "mattermost"

func TestServer(c *model.Client4, l *log.Logger) error {
	if props, resp := c.GetOldClientConfig(""); resp.Error != nil {
		// server ping failed!
		return resp.Error
	} else {
		l.Printf("Server detected and is running version " + props["Version"])
		l.Printf("Server returned: %#v\n", props)
		return nil
	}
}

func main() {

	var (
		l   *log.Logger
		err error
	)

	// parse command line params
	cmd, err := cmdline.NewCmdArgs(app)
	if err != nil {
		log.Fatalf("ERROR: parse command line failed with: %q\n", err)
	}

	if *(cmd.Version) {
		l = log.New(os.Stderr, app+":", log.Lshortfile|log.Ldate|log.Lmicroseconds)
		version.DumpVersion(l)
		os.Exit(0)
	}

	// read config file
	cfg := config.NewConfigContainer()
	err = cfg.GetConfig(*(cmd.CfgFile))
	if err != nil {
		log.Fatalf("ERROR: getConfig() failed: %s\n", err)
	}

	// setup logger
	l, err = logger.SetupLogger(cfg.Logfile, app)
	if err != nil {
		log.Fatalf("setuplogger(%s) failed with: %q\n", cfg.Logfile, err)
	}

	l.Println("==================== mattermostbot start =======================================")

	version.DumpVersion(l)

	if _, ok := cfg.Connectors[app]["triggerString"]; !ok {
		l.Printf("Info: fetching config[connectors][mattermost][triggerString] failed: setting it to default: showmeslurm\n")
		cfg.Connectors[app]["triggerString"] = "showmeslurm"
	}
	if _, ok := cfg.Connectors[app]["token"]; !ok {
		l.Fatalf("MAIN: fetching config[connectors][mattermost][token] failed: %s\n", err)
	}

	l.Printf("Starting: %q\n", cfg.Connectors[app]["name"])

	client := model.NewAPIv4Client(cfg.Connectors[app]["serverurl"])
	client.SetOAuthToken(cfg.Connectors[app]["token"])
	l.Printf("\nclient: %#v\n", client)

	if e := TestServer(client, l); e != nil {
		l.Fatalf("Can not proceed, TestServer() returned error: %s\n", e)
	}

	// Main loop
	for {
		webSocketClient, err := model.NewWebSocketClient4(cfg.Connectors[app]["wsurl"], client.AuthToken)
		if err != nil {
			l.Fatalf("ERROR: NewWebSocketClient(): %s\n", err)
		}
		l.Printf("Connected to WS: %s\n", cfg.Connectors[app]["wsurl"])
		l.Printf("Websocketclient: %#v\n\n", webSocketClient)
		webSocketClient.Listen()

		l.Printf("Listening to event channels...\n")
		for resp := range webSocketClient.EventChannel {
			l.Printf("GOT WS EVENT: %#v\n", resp)
			data := resp.GetData()
			l.Printf("Data: %#v\n", data)
			l.Printf("Channel name: %s\n", data["channel_name"])
			l.Printf("JSON: %s\n", resp.ToJson())

			x, ok := resp.GetData()["post"].(string)
			if !ok {
				l.Printf("Info: post == nil, skipping.\n")
			} else {
				post := model.PostFromJson(strings.NewReader(x))
				l.Printf("POST: %#v\n", post)
				l.Printf("POST.channelid: %s\n", post.ChannelId)
				l.Printf("POST.userid: %s\n", post.UserId)
				l.Printf("POST.message: %s\n", post.Message)
				if strings.Contains(post.Message, cfg.Connectors[app]["triggerString"]) {
					// Post something back!
					resPost := model.Post{}
					resPost.ChannelId = post.ChannelId
					resPost.Message = fmt.Sprintf("Hello!\nI'm %s!\nTo receive your job results here, use the following switch in your job scripts:\n--mail-user=mm:%s\n", cfg.Connectors[app]["name"], resPost.ChannelId)
					if _, r := client.CreatePost(&resPost); r.Error != nil {
						l.Printf("Post response to chan: %s successfull!\n", resPost.ChannelId)
					} else {
						l.Printf("Post response FAILED!\n")
					}
				}
			}
		}
	}

	l.Println("==================== mattermostbot end =========================================")
}
