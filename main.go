package main

import (
	"context"
	"flag"
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/fsnotify/fsnotify"
	log "github.com/sirupsen/logrus"
)

var (
	Token        string
	DenylistPath string
	lock         sync.Mutex
	denylist     []*regexp.Regexp
)

func main() {
	flag.StringVar(&Token, "token", "", "Discord bot token")
	flag.StringVar(&DenylistPath, "denylist", "", "Filepath to denylist of regular expressions, separated by new line delimiters")
	flag.Parse()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	// Create a new Discord session using the provided bot token.
	dg, err := discordgo.New("Bot " + Token)
	if err != nil {
		fmt.Println("error creating Discord session,", err)
		return
	}
	// Register the messageCreate func as a callback for MessageCreate events.
	dg.AddHandler(messageCreate)

	// Monitor denylist changes
	go monitorDenylistFile(ctx, DenylistPath)

	// Open a websocket connection to Discord and begin listening.
	err = dg.Open()
	if err != nil {
		fmt.Println("error opening connection,", err)
		return
	}
	defer func() {
		if err := dg.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	// Wait here until CTRL-C or other term signal is received.
	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM)
	<-sc
}

// This function will be called (due to AddHandler above) every time a new
// message is created on any channel that the autenticated bot has access to.
func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	// Ignore all messages created by the bot itself
	if m.Author.ID == s.State.User.ID {
		return
	}
	for _, re := range denylist {
		if re.MatchString(m.Content) {
			if err := s.ChannelMessageDelete(m.ChannelID, m.ID); err != nil {
				log.WithError(err).Error("Failed to delete denied message.")
			}
			ts, err := discordgo.SnowflakeTimestamp(m.Author.ID)
			if err != nil {
				log.WithError(err).Error("Could not determine user's timestamp")
			}
			age := time.Since(ts)
			log.WithFields(log.Fields{
				"username":   m.Author.Username,
				"id":         m.Author.ID,
				"content":    m.Content,
				"accountAge": age,
				"regexp":     re.String(),
			}).Info("Message deleted")
		}
	}
}

func monitorDenylistFile(ctx context.Context, fp string) {
	updateDenyList(fp)
	log.WithField("filepath", fp).Info("Monitoring denylist for file changes")

	w, err := fsnotify.NewWatcher()
	if err != nil {
		log.WithError(err).Error("Failed to create fsnotify watcher")
		return
	}
	if err := w.Add(fp); err != nil {
		log.WithError(err).Error("Failed to create fsnotify watcher")
		return
	}
	for {
		select {
		case <-w.Events:
			lock.Lock()
			updateDenyList(fp)
			lock.Unlock()
		case <-ctx.Done():
			return
		}
	}
}

func updateDenyList(fp string) {
	newDenyList := make([]*regexp.Regexp, 0)
	content, err := ioutil.ReadFile(fp)
	if err != nil {
		log.WithError(err).Error("Failed to read denylist")
		return
	}
	s := string(content)
	for _, row := range strings.Split(s, "\n") {
		if row == "" {
			continue
		}
		re, err := regexp.Compile("(?i)" + row) // Prefix (?i) to make case insenstive.
		if err != nil {
			log.WithError(err).Errorf("Failed to parse regex: %s", row)
			continue
		}
		newDenyList = append(newDenyList, re)
	}

	if len(newDenyList) > 0 {
		denylist = newDenyList
		log.WithField("count", len(newDenyList)).Info("Updated deny list")
	}
}