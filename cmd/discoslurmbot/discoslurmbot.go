package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/CLIP-HPC/goslmailer/internal/cmdline"
	"github.com/CLIP-HPC/goslmailer/internal/config"
	"github.com/CLIP-HPC/goslmailer/internal/logger"
	"github.com/CLIP-HPC/goslmailer/internal/version"
	"github.com/bwmarrin/discordgo"
)

const app = "discoslurmbot"

type botConfig struct {
	config.ConfigContainer
	l *log.Logger
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the authenticated bot has access to.
//
// It is called whenever a message is created but only when it's sent through a
// server as we did not request IntentsDirectMessages.
func messageCreate(bc botConfig) func(*discordgo.Session, *discordgo.MessageCreate) {
	return func(s *discordgo.Session, m *discordgo.MessageCreate) {

		fmt.Printf("session: %#v\n", s)
		fmt.Printf("message: %#v\n", m)
		fmt.Printf("message content: %#v\n", m.Content)
		fmt.Printf("author: %#v\n", m.Author.ID)
		fmt.Printf("user.id: %#v\n", s.State.User.ID)

		// Ignore all messages created by the bot itself
		// This isn't required in this specific example but it's a good practice.
		if m.Author.ID == s.State.User.ID {
			return
		}
		// In this example, we only care about messages that are "ping".
		if m.Content != bc.Connectors["discord"]["triggerString"] {
			return
		}

		// We create the private channel with the user who sent the message.
		channel, err := s.UserChannelCreate(m.Author.ID)
		if err != nil {
			// If an error occurred, we failed to create the channel.
			//
			// Some common causes are:
			// 1. We don't share a server with the user (not possible here).
			// 2. We opened enough DM channels quickly enough for Discord to
			//    label us as abusing the endpoint, blocking us from opening
			//    new ones.
			bc.l.Println("error creating channel:", err)
			s.ChannelMessageSend(
				m.ChannelID,
				"Something went wrong while sending the DM!",
			)
			return
		}
		// Then we send the message through the channel we created.
		msg := fmt.Sprintf("Welcome,\nI am %s,\nplease use this switch in your job submission script in addition to '--mail-type' and i'll get back to you:\n '--mail-user=discord:%s'", bc.Connectors["discord"]["botname"], channel.ID)
		_, err = s.ChannelMessageSend(channel.ID, msg)
		if err != nil {
			// If an error occurred, we failed to send the message.
			//
			// It may occur either when we do not share a server with the
			// user (highly unlikely as we just received a message) or
			// the user disabled DM in their settings (more likely).
			bc.l.Println("error sending DM message:", err)
			s.ChannelMessageSend(
				m.ChannelID,
				"Failed to send you a DM. "+
					"Did you disable DM in your privacy settings?",
			)
		}
	}
}

func main() {

	// parse command line params
	cmd, err := cmdline.NewCmdArgs(app)
	if err != nil {
		log.Fatalf("ERROR: parse command line failed with: %q\n", err)
	}

	if *(cmd.Version) {
		l := log.New(os.Stderr, app+":", log.Lshortfile|log.Ldate|log.Lmicroseconds)
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
	l, err := logger.SetupLogger(cfg.Logfile, "gobler")
	if err != nil {
		log.Fatalf("setuplogger(%s) failed with: %q\n", cfg.Logfile, err)
	}

	l.Println("===================== discoslurmbot start ======================================")

	version.DumpVersion(l)

	if _, ok := cfg.Connectors["discord"]["token"]; !ok {
		l.Fatalf("MAIN: fetching config[connectors][discord][token] failed: %s\n", err)
	}

	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + cfg.Connectors["discord"]["token"])
	if err != nil {
		l.Println("error creating Discord session,", err)
		return
	}

	// Register the messageCreate func as a callback for MessageCreate events.
	bc := botConfig{
		*cfg,
		l,
	}
	dg.AddHandler(messageCreate(bc))

	// In this example, we only care about receiving message events.
	// pja: and DMs
	dg.Identify.Intents = discordgo.IntentsGuildMessages | discordgo.IntentDirectMessages

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		l.Println("error opening connection,", err)
		return
	}

	// Wait here until CTRL-C or other term signal is received.
	l.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	dg.Close()

	l.Println("===================== discoslurmbot end ========================================")
}
