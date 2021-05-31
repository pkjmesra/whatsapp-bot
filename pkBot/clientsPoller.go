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
	globalWACClient *pkWhatsApp.WhatsappClient
)

const (
	defaultSearchInterval 	= 60
)

func SetWACClient(wac *pkWhatsApp.WhatsappClient) {
	globalWACClient = wac
}

func Clients() map[string]*RemoteClient{
	if remoteClients == nil{
		remoteClients = make(map[string]*RemoteClient)
		allUsers, _ := readUsers()
		for _, subscriber := range allUsers.Subscribers {
			remoteClients[subscriber.RemoteJID] = NewRemoteClient(pkWhatsApp.Message{}, subscriber.RemoteJID, subscriber.MobileNumber, globalWACClient)
		}
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

func subscribe(remoteClient *RemoteClient) {
	allUsers, _ := readUsers()
	newSubscriber := Subscriber{RemoteJID: remoteClient.RemoteJID,
								MobileNumber : remoteClient.RemoteMobileNumber,
								Date : time.Now().String(),
							}
	allUsers.Subscribers = append(allUsers.Subscribers, newSubscriber)
	count := 0
	for _, unsubscriber := range allUsers.NonSubscribers {
		if unsubscriber.RemoteJID == remoteClient.RemoteJID {
			allUsers.NonSubscribers = removeSubscriber(allUsers.NonSubscribers, count)
			break
		}
		count = count + 1
	}
	writeUsers(allUsers)
	var cmd = evaluateInput(remoteClient, "subscribe")
	remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ToBeSent)
}

func unsubscribe(remoteClient *RemoteClient) {
	allUsers, _ := readUsers()
	subs := Subscriber{}
	count := 0
	for _, subscriber := range allUsers.Subscribers {
		if subscriber.RemoteJID == remoteClient.RemoteJID {
			allUsers.Subscribers = removeSubscriber(allUsers.Subscribers, count)
			subs = subscriber
			break
		}
		count = count + 1
	}
	allUsers.NonSubscribers = append(allUsers.NonSubscribers, subs)
	writeUsers(allUsers)
	var cmd = evaluateInput(remoteClient, "unsubscribe")
	remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ToBeSent)
}

func CheckSlots(remoteClient *RemoteClient) error {
	// Search for slots
	if remoteClient == nil {
		fmt.Println(remoteClient.RemoteJID + ":RemoteClient is nil")
		return nil
	}
	if remoteClient.LastSent != nil && remoteClient.LastSent.Name == "otp" {
		Search(remoteClient)
		remoteClient.PollWaitCounter = remoteClient.PollWaitCounter + 1
		return nil
	}
	bk, _ := Search(remoteClient)
	if bk.Description != "" { // || debug
		fmt.Println(remoteClient.RemoteJID + ":Found available slot. Delaying polling now.")
		resetPollingInterval(remoteClient, defaultSearchInterval)
		bk.BookAnySlot = remoteClient.Params.BookingPrefs.BookAnySlot
		remoteClient.Params.BookingPrefs = bk
		writeUser(remoteClient)
		var cmd = evaluateInput(remoteClient, "search")
		remoteClient.LastSent = cmd
		askUserForOTP(remoteClient, cmd)
	} else {
		fmt.Println(remoteClient.RemoteJID + ":No results from search")
	}
	return nil
}

func resetPollingInterval(remoteClient *RemoteClient, interval int) {
	remoteClient.PollingInterval = interval
	fmt.Fprintf(os.Stderr,remoteClient.RemoteJID + ":Restarting ticker with interval: %d seconds\n", interval)
	ticker = time.NewTicker(time.Second * time.Duration(interval))
}
