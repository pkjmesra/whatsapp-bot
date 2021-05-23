package main

import (
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/pkjmesra/whatsapp-bot/pkBot"
	"github.com/pkjmesra/whatsapp-bot/pkWhatsApp"
	// "github.com/pkg/errors"
	"github.com/spf13/cobra"
)

var (
	remoteMobile string
	thisRemoteClient *pkBot.RemoteClient
	interval int
	rootCmd = &cobra.Command{
		Use:   "cwhatsapp-bot [FLAGS]",
		Short: "CoWIN Vaccine availability notifier India",
		RunE: func(cmd *cobra.Command, args []string) error {
			return Run(args)
		},
	}
)

const (
	remoteMobileEnv        	= "REMOTE_MOBILE"
	searchIntervalEnv 		= "SEARCH_INTERVAL"
	defaultSearchInterval 	= 60
)

func init() {
	rootCmd.PersistentFlags().StringVarP(&remoteMobile, "remoteMobile", "m", os.Getenv(remoteMobileEnv), "Remote Mobile number")
	rootCmd.PersistentFlags().IntVarP(&interval, "interval", "i", getIntEnv(searchIntervalEnv), fmt.Sprintf("Interval to repeat the search. Default: (%v) second", defaultSearchInterval))
}

// Execute executes the main command
func Execute() error {
	fmt.Println("Now executing!")
	return rootCmd.Execute()
}

func main() {
	remoteClients := make(map[string]*pkBot.RemoteClient)
	client := pkWhatsApp.NewClient()
	pkBot.Initialize()
	client.Listen(func(msg pkWhatsApp.Message) {
		// Only handle messages to self or one-on-one messages to the registered WhatsApp number
		if strings.Contains(msg.From, "@c.us") || strings.Contains(msg.From, "@s.whatsapp.net") {
			addNewRemoteClient(remoteClients, msg, client)
		} else if strings.Contains(msg.From, "@g.us"){
			// Ignore the messages in group
			fmt.Println("Message Received -> ID : " + msg.From + " : Message:" + msg.Text)
		}
	})
	Execute()
}

func addNewRemoteClient(m map[string]*pkBot.RemoteClient, msg pkWhatsApp.Message, wac *pkWhatsApp.WhatsappClient) *pkBot.RemoteClient {
    rc := m[msg.From]
    if rc == nil {
        m[msg.From] = pkBot.NewClient(msg, wac)
        rc = m[msg.From]
    }
    rc.Received = msg
    rc.RemoteJID = remoteMobile + "@s.whatsapp.net"
    rc.RemoteMobileNumber = remoteMobile
    fmt.Println("RemoteJID set to:" + rc.RemoteJID)
    pkBot.Respond(rc)
    thisRemoteClient = rc
    return rc
}

func getIntEnv(envVar string) int {
	v := os.Getenv(envVar)
	if len(v) == 0 {
		return 0
	}
	i, err := strconv.Atoi(v)
	if err != nil {
		log.Fatal(err)
	}
	return i
}

func Run(args []string) error {
	// always make a log file
	logfile, err := os.OpenFile("run.log", os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatalf("error opening a log file: %v", err)
	}
	defer logfile.Close()
	log.SetOutput(logfile)

	if err := checkSlots(); err != nil {
		return err
	}
	ticker := time.NewTicker(time.Second * time.Duration(interval))
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			if err := checkSlots(); err != nil {
				return err
			}
		}
	}
	return nil
}

func checkSlots() error {
	// Search for slots
	if thisRemoteClient == nil {
		return nil
	}
	bk, err := pkBot.Search(thisRemoteClient)
	if bk.Description != "" { // || debug 
		pkBot.SendResponse(thisRemoteClient, "book")
	}
	return err
}