package pkBot

import (
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	"github.com/pkjmesra/whatsapp-bot/pkWhatsApp"
)

var (
	remoteClients map[string]*RemoteClient
	globalPollingInterval int
	lastPollingInterval int
	pollingRemoteClient *RemoteClient
	ticker *time.Ticker
	bookingInProgress bool
)

const (
	defaultSearchInterval 	= 60
)
func Clients() map[string]*RemoteClient{
	if remoteClients == nil{
		remoteClients = make(map[string]*RemoteClient)
	}
	return remoteClients
}

func GetClient(msg pkWhatsApp.Message, wac *pkWhatsApp.WhatsappClient) *RemoteClient {
	clients := Clients()
	rc := clients[msg.From]
	if rc == nil {
        remoteClients[msg.From] = NewClient(msg, wac)
        rc = Clients()[msg.From]
        rmtMobile := strings.TrimSuffix(string(msg.From), "@s.whatsapp.net")
        rc.RemoteJID = msg.From
    	rc.RemoteMobileNumber = rmtMobile
    	rc.Received = msg
    	fmt.Println("RemoteJID set to:" + rc.RemoteJID)
    }
    return rc
}

func BeginPollingForClients(interval int) error {
	globalPollingInterval = interval
	for _, rc := range Clients() {
		clntExists := allSubscribersMap[rc.RemoteJID]
		if clntExists != nil {
			PollServer(rc)
		}
	}
	return nil
}

func PollServer(remoteClient *RemoteClient) error {
	if remoteClient == nil {
		return nil
	}
	if remoteClient.PollingInterval == 0 {
		remoteClient.PollingInterval = globalPollingInterval
	}
	fmt.Fprintf(os.Stderr, "Polling every %d seconds\n", remoteClient.PollingInterval)
	ticker = time.NewTicker(time.Second * time.Duration(remoteClient.PollingInterval))
	for {
		select {
		case <-ticker.C:
			if remoteClient.PollingInterval == defaultSearchInterval && remoteClient.PollWaitCounter >= 3 {
				resetPollingInterval(remoteClient, globalPollingInterval)
				remoteClient.PollWaitCounter = 0
			}
			fmt.Fprintf(os.Stderr, "Ticker.Tick. Polling every %d seconds\n", globalPollingInterval)
			for _, rc := range Clients() {
				clntExists := allSubscribersMap[rc.RemoteJID]
				if clntExists != nil {
					if err := CheckSlots(rc); err != nil {
						log.Fatalf(rc.RemoteJID + ":Error in CheckSlots(PollServer): %v\n", err)
						return nil
					}
				}
			}
		}
	}
	fmt.Println("Stopped ticker.")
	return nil
}
