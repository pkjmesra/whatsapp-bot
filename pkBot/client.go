package pkBot

import (
	"fmt"
	"os"
	"strings"

	"github.com/pkjmesra/whatsapp-bot/pkWhatsApp"
)

// NewClient create whatsapp client
func NewClient(msg pkWhatsApp.Message, wac *pkWhatsApp.WhatsappClient) *RemoteClient {
	mob := strings.TrimSuffix(string(msg.From), "@c.us")
	mob = strings.TrimSuffix(string(mob), "@s.whatsapp.net")
	rc := NewRemoteClient(msg, msg.From, mob, wac)
	return rc
}

func NewRemoteClient(msg pkWhatsApp.Message, from, mobile string, wac *pkWhatsApp.WhatsappClient) *RemoteClient {
	rc := RemoteClient {
							RemoteJID: from,
							RemoteMobileNumber: mobile,
							Host: wac,
							PollingInterval: 0,
							PollWaitCounter: 0,
							Received: msg,
							LastSent: &Command{},
							LastReceived: pkWhatsApp.Message{},
							Params: &UserParams{},
						}
	resetParams(&rc, true)
	return &rc
}

func Respond (remoteClient *RemoteClient) {
	// message := remoteClient.Received.Source
	userInput := strings.ToLower(remoteClient.Received.Text)
	if userInput == "subscribe" {
		subscribe(remoteClient)
	} else if userInput == "unsubscribe" {
		unsubscribe(remoteClient)
	}
	clntExists := allSubscribersMap[remoteClient.RemoteJID]
	if clntExists == nil {
		return
	}
	lastSent := remoteClient.LastSent
	if userInput == "vaccine" || userInput == "book" {
		remoteClient.Params, _ = readUser(remoteClient)
		SendResponse(remoteClient, userInput)
	} else if userInput == "certificate" {
		SendResponse(remoteClient, userInput)
	} else if lastSent.Name == "" {
		fmt.Println(remoteClient.RemoteJID + ":Restarting because lastsent.Name is empty")
		SendResponse(remoteClient, "")
	} else {
		saveUserInput(remoteClient)
		processUserInput(remoteClient)
	}
}

func SendResponse(remoteClient *RemoteClient, userInput string) {
	fmt.Println(remoteClient.RemoteJID + ":Received Response request with userInput:" + userInput)
	if userInput == "" {
		resetPollingInterval(remoteClient, globalPollingInterval)
	}
	var cmd = evaluateInput(remoteClient, userInput)
	if cmd.CommandType == "UserInput" {
		handleUserInputs(remoteClient, cmd)
	} else if cmd.CommandType == "Function" {
		handleFunctionCommands(remoteClient, cmd)
	} else {
		fmt.Println(remoteClient.RemoteJID + ":Restarting because commandtype is neither function nor userInput")
		SendResponse(remoteClient, "")
	}
}

func handleFunctionCommands(remoteClient *RemoteClient, cmd *Command) {
	lastSent := remoteClient.LastSent
	if cmd.Name == "book" {
		handleAdhocBookingRequest(remoteClient, cmd)
	}
	if lastSent.NextCommand == "search" {
		handleSearchRequest(remoteClient, cmd)
	} else if cmd.Name == "otpbeneficiary" {
		generateOTPForBearerToken(remoteClient, cmd)
	} else if cmd.Name == "beneficiariesupdate" {
		getBeneficiariesForRemoteUser(remoteClient, cmd, true)
	} else if lastSent.NextCommand == "beneficiaries" {
		getBeneficiariesForRemoteUser(remoteClient, cmd, false)
	} else if lastSent.NextCommand == "readcaptcha" {
		sendCAPTCHA(remoteClient, cmd)
		updateClient(remoteClient, cmd)
	} else if lastSent.NextCommand == "bookingconfirmation" {
		handleBookingRequest(remoteClient, cmd)
	} else if cmd.Name == "certificate" {
		generateOTPForBearerToken(remoteClient, cmd)
	} else if cmd.Name == "downloadcertificate" {
		downloadcertificate(remoteClient, cmd)
	}
}

func resetParams(remoteClient *RemoteClient, shouldReload bool) {
	fmt.Println(remoteClient.RemoteJID + ":Resetting Params.")
	bookingCenterId = 0
	remoteClient.Params = &UserParams{}
	remoteClient.Params.CAPTCHA = ""
	remoteClient.Params.OTPTxnDetails = &OTPTxn{}
	remoteClient.Params.BookingPrefs = &BookingSlot{CenterID:0, BookAnySlot:true}
	remoteClient.Params.Beneficiaries = &BeneficiaryList{}
	if shouldReload {
		remoteClient.Params, _ = readUser(remoteClient)
		remoteClient.Params.OTPTxnDetails = &OTPTxn{}
	} else {
		writeUser(remoteClient)
	}
}

func updateClient(remoteClient *RemoteClient, cmd *Command) {
	remoteClient.LastSent = cmd
	remoteClient.LastReceived = remoteClient.Received
	remoteClient.Received = pkWhatsApp.Message{}
	fmt.Fprintf(os.Stderr, remoteClient.RemoteJID + ":LastSent: \n{Name: %s, NextCmd: %s}\n", cmd.Name, cmd.NextCommand)
}
