package pkBot

import (
	"strings"
	"strconv"
	"github.com/pkjmesra/whatsapp-bot/pkWhatsApp"
)

// Remote WhatsappClient who's trying to communicate
type RemoteClient struct {
	RemoteJID 			string
	LastReceived		pkWhatsApp.Message
	LastSent			*Command
	Received 			pkWhatsApp.Message
	Host				*pkWhatsApp.WhatsappClient
	Params 	 			SearchParams
}

type SearchParams struct {
	State 			string
	District		string
	Age				int
	OTP 			int
}
// NewClient create whatsapp client
func NewClient(msg pkWhatsApp.Message, wac *pkWhatsApp.WhatsappClient) *RemoteClient {
	rc := RemoteClient {
							RemoteJID: msg.From, 
							Host: wac, 
							Received: msg, 
							LastSent: &Command{}, 
							LastReceived: pkWhatsApp.Message{},
							Params: SearchParams{},
						}
	bookingCenterId = 0
	return &rc
}

func Respond (remoteClient *RemoteClient) {
	message := remoteClient.Received.Source
	if message.Info.FromMe {
		lastSent := remoteClient.LastSent
		if lastSent.Name == "" {
			sendResponse(remoteClient, "")
		} else {
			saveUserInput(remoteClient)
			processUserInput(remoteClient)
		}
	} else {
		remoteClient.Host.SendText(remoteClient.RemoteJID, "Hello from *github*!")
	}
}

func sendResponse(remoteClient *RemoteClient, userInput string) {
	var cmd = evaluateInput(userInput)
	if cmd.CommandType == "UserInput" {
		remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ToBeSent)
		remoteClient.LastSent = cmd
		remoteClient.LastReceived = remoteClient.Received
		remoteClient.Received = pkWhatsApp.Message{}
	} else if cmd.CommandType == "Function" {
		lastSent := remoteClient.LastSent
		if lastSent.NextCommand == "search" {
			var err error
			var bk *BookingSlot
			params := remoteClient.Params
			bk, err = searchByStateDistrict(params.Age, params.State, params.District)
			if err != nil{
				sendResponse(remoteClient, "")
			} else {
				if bk.Description == "" {
					cmd.ToBeSent = cmd.ErrorResponse
				}
				remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ToBeSent + bk.Description)
				remoteClient.LastSent = cmd
				remoteClient.LastReceived = remoteClient.Received
				remoteClient.Received = pkWhatsApp.Message{}
				if bk.Description == "" {
					sendResponse(remoteClient, "")
				}
			}
		} else if lastSent.NextCommand == "beneficiaries" {
			
		}
	} else {
		sendResponse(remoteClient, "")
	}
}

func processUserInput(remoteClient *RemoteClient) {
	var userInput = remoteClient.Received.Text
	var lastSent = remoteClient.LastSent
	
	if lastSent.ExpectedResponse != "" {
		if lastSent.ExpectedResponse == strings.ToLower(userInput) {
			sendResponse(remoteClient, lastSent.NextCommand)
		} else {
			sendResponse(remoteClient, "")
		}
	} else {
		sendResponse(remoteClient, lastSent.NextCommand)
	}
}

func saveUserInput(remoteClient *RemoteClient) {
	var userInput = remoteClient.Received.Text
	var lastSent = remoteClient.LastSent
	if lastSent.Name == "vaccine" {
		remoteClient.Params.State = userInput
	} else if lastSent.Name == "district" {
		remoteClient.Params.District = userInput
	} else if lastSent.Name == "age" {
		if age, err := strconv.Atoi(userInput); err == nil {
			remoteClient.Params.Age = age
		}
	} else if lastSent.Name == "otp" {
		if otp, err := strconv.Atoi(userInput); err == nil {
			remoteClient.Params.OTP = otp
		}
	}
}