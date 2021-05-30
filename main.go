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
	globalPollingInterval int
	globalTicker *time.Ticker
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
	rootCmd.PersistentFlags().IntVarP(&globalPollingInterval, "interval", "i", getIntEnv(searchIntervalEnv), fmt.Sprintf("Interval to repeat the search. Default: (%v) second", defaultSearchInterval))
}

// Execute executes the main command
func Execute() error {
	fmt.Println("Now executing!")
	return rootCmd.Execute()
}

func main() {
	client := pkWhatsApp.NewClient()
	pkBot.Initialize()
	client.Listen(func(msg pkWhatsApp.Message) {
		// Only handle messages to self or one-on-one messages to the registered WhatsApp number
		if strings.Contains(msg.From, "@c.us") || strings.Contains(msg.From, "@s.whatsapp.net") {
			addNewRemoteClient(msg, client)
		} else if strings.Contains(msg.From, "@g.us"){
			// Ignore the messages in group
			fmt.Println("Message Received (and ignored) -> ID : " + msg.From + " : Message:" + msg.Text)
		}
	})
	Execute()
}

func addNewRemoteClient(msg pkWhatsApp.Message, wac *pkWhatsApp.WhatsappClient) *pkBot.RemoteClient {
    remoteClient := pkBot.GetClient(msg, wac)
    remoteClient.Received = msg
    pkBot.Respond(remoteClient)
    return remoteClient
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

	clients := pkBot.Clients()
	prevCount := len(clients)
	pkBot.BeginPollingForClients(globalPollingInterval)
	globalTicker = time.NewTicker(time.Second * time.Duration(defaultSearchInterval))
	for {
		select {
		case <-globalTicker.C:
			newClients := pkBot.Clients()
			newCount := len(newClients)
			if prevCount != newCount {
				// New client got added or old client got removed
				fmt.Fprintf(os.Stderr, "GlobalTicker.Tick. Polling for %d client now\n", len(clients))
				pkBot.BeginPollingForClients(globalPollingInterval)
			}
		}
	}
	fmt.Println("Stopped Global ticker.")
	return nil
}
