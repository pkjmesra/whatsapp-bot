package pkBot

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	
	"github.com/pkjmesra/whatsapp-bot/pkWhatsApp"
)

// Remote WhatsappClient who's trying to communicate
type RemoteClient struct {
	RemoteJID 			string
	RemoteMobileNumber	string
	LastReceived		pkWhatsApp.Message
	LastSent			*Command
	Received 			pkWhatsApp.Message
	Host				*pkWhatsApp.WhatsappClient
	Params 	 			*UserParams
}

type UserParams struct {
	State 			string
	District		string
	Age				int
	OTP 			int
	OTPTxnDetails 	*OTPTxn
	Beneficiaries	*BeneficiaryList
	BookingPrefs  	*BookingSlot
	CAPTCHA 		string
	ConfirmationID 	string
}
// NewClient create whatsapp client
func NewClient(msg pkWhatsApp.Message, wac *pkWhatsApp.WhatsappClient) *RemoteClient {
	mob := strings.TrimSuffix(string(msg.From), "@c.us")
	mob = strings.TrimSuffix(string(mob), "@s.whatsapp.net")
	rc := RemoteClient {
							RemoteJID: msg.From,
							RemoteMobileNumber: mob,
							Host: wac,
							Received: msg,
							LastSent: &Command{},
							LastReceived: pkWhatsApp.Message{},
							Params: &UserParams{},
						}
	resetParams(&rc)
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
		// remoteClient.Host.SendText(remoteClient.RemoteJID, "Hello from *github*!")
	}
}

func sendResponse(remoteClient *RemoteClient, userInput string) {
	var cmd = evaluateInput(remoteClient, userInput)
	var err error
	if cmd.CommandType == "UserInput" {
		remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ToBeSent)
		updateClient(remoteClient, cmd)
	} else if cmd.CommandType == "Function" {
		lastSent := remoteClient.LastSent
		if lastSent.NextCommand == "search" {
			var bk *BookingSlot
			params := remoteClient.Params
			bk, err = searchByStateDistrict(params.Age, params.State, params.District)
			if err != nil{
				fmt.Println("Error while searching for slots")
				sendResponse(remoteClient, "")
			} else {
				if bk.Description == "" {
					fmt.Println("No slot found from the search")
					cmd.ToBeSent = cmd.ErrorResponse1
					remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ToBeSent)
					sendResponse(remoteClient, "")
					return
				}
				if bookingCenterId <= 0 {
					fmt.Println("All slots being shared with user since preferred booking center is not set")
					remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ToBeSent + bk.Description)
				}
				if bookingCenterId > 0 && bk.Description != "" {
					fmt.Println("Preferred booking center is set already and center has slots")
					remoteClient.Params.BookingPrefs = bk
					var txn *OTPTxn
					txn, err = generateOTP(remoteClient.RemoteMobileNumber, false)
					if err == nil && txn.TXNId != "" {
						remoteClient.Params.OTPTxnDetails = txn
						writeUser(remoteClient)
						updateClient(remoteClient, cmd)
						sendResponse(remoteClient, cmd.NextCommand)
						return
					} else {
						fmt.Println("Failed generating OTP for " + remoteClient.RemoteMobileNumber)
						remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ErrorResponse2)
						sendResponse(remoteClient, "")
						return
					}
				}
				updateClient(remoteClient, cmd)
			}
		} else if lastSent.NextCommand == "beneficiaries" {
			var beneficiariesList *BeneficiaryList
			beneficiariesList, err = getBeneficiaries()
			if err == nil && beneficiariesList.Description != ""{
				remoteClient.Params.Beneficiaries = beneficiariesList
				writeUser(remoteClient)
				cmd.ToBeSent = fmt.Sprintf(cmd.ToBeSent, beneficiariesList.Description)
				remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ToBeSent)
				updateClient(remoteClient, cmd)
				sendResponse(remoteClient, cmd.NextCommand)
			}
		} else if lastSent.NextCommand == "readCAPTCHA" {
			sendCAPTCHA(remoteClient, cmd)
			updateClient(remoteClient, cmd)
		} else if lastSent.NextCommand == "bookingConfirmation" {
			var cnf string
			cnf, err = bookAppointment(remoteClient.Params.Beneficiaries, remoteClient.Params.CAPTCHA, remoteClient.Params.BookingPrefs)
			if err == nil {
				if remoteClient.Params.ConfirmationID == "" {
					remoteClient.Params.ConfirmationID = cnf
				} else {
					remoteClient.Params.ConfirmationID = remoteClient.Params.ConfirmationID + "," + cnf
				}
				writeUser(remoteClient)
				cmd.ToBeSent = fmt.Sprintf(cmd.ToBeSent, cnf)
				remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ToBeSent)
				updateClient(remoteClient, cmd)
			} else {
				cmd.ErrorResponse1 = fmt.Sprintf(cmd.ErrorResponse1, err.Error())
				remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ErrorResponse1)
				sendResponse(remoteClient, "")
			}
		}
	} else {
		sendResponse(remoteClient, "")
	}
}

func sendOTP(remoteClient *RemoteClient, cmd *Command) {
	txn, err := generateOTP(remoteClient.RemoteMobileNumber, false)
	if err == nil && txn.TXNId != "" {
		remoteClient.Params.OTPTxnDetails = txn
		writeUser(remoteClient)
		updateClient(remoteClient, cmd)
		sendResponse(remoteClient, cmd.NextCommand)
		return
	} else {
		fmt.Println("Failed generating OTP for " + remoteClient.RemoteMobileNumber)
		remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ErrorResponse2)
		sendResponse(remoteClient, "")
		return
	}
}

func sendCAPTCHA(remoteClient *RemoteClient, cmd *Command) {
	fileName := remoteClient.RemoteMobileNumber + "_captcha"
	err := getCaptchaSVG(fileName + ".svg")
	var f string
	if err != nil {
		fmt.Println("Failed getting CAPTCHA for " + remoteClient.RemoteMobileNumber)
		remoteClient.Host.SendText(remoteClient.RemoteJID, fmt.Sprintf(cmd.ErrorResponse1, remoteClient.RemoteMobileNumber))
		sendResponse(remoteClient, "")
		return
	}
	err = exportToPng(fileName + ".svg", fileName + ".png")
	if err != nil {
		fmt.Println("Failed getting CAPTCHA for " + remoteClient.RemoteMobileNumber)
		remoteClient.Host.SendText(remoteClient.RemoteJID, fmt.Sprintf(cmd.ErrorResponse1, remoteClient.RemoteMobileNumber))
		sendResponse(remoteClient, "")
		return
	}
	_ , f, _ , err = getPngImage(fileName + ".png")
	if err != nil {
		fmt.Println("Failed getting CAPTCHA for " + remoteClient.RemoteMobileNumber)
		remoteClient.Host.SendText(remoteClient.RemoteJID, fmt.Sprintf(cmd.ErrorResponse1, remoteClient.RemoteMobileNumber))
		sendResponse(remoteClient, "")
		return
	}
	f, err = getJPEGImage(fileName + ".png")
	if err != nil {
		fmt.Println("Failed getting CAPTCHA for " + remoteClient.RemoteMobileNumber)
		remoteClient.Host.SendText(remoteClient.RemoteJID, fmt.Sprintf(cmd.ErrorResponse1, remoteClient.RemoteMobileNumber))
		sendResponse(remoteClient, "")
		return
	}
	remoteClient.Host.SendImage(remoteClient.RemoteJID, f, "jpeg", cmd.ToBeSent)
}

func processUserInput(remoteClient *RemoteClient) {
	var userInput = remoteClient.Received.Text
	var lastSent = remoteClient.LastSent
	var err error
	if lastSent.ExpectedResponse != "" {
		if lastSent.ExpectedResponse == strings.ToLower(userInput) {
			sendResponse(remoteClient, lastSent.NextCommand)
		} else {
			sendResponse(remoteClient, "")
		}
	} else {
		userInput = strings.ToLower(userInput)
		nextCmd := lastSent.NextCommand
		if (userInput == "y" || userInput == "yes") && lastSent.NextYCommand != "" {
			nextCmd = lastSent.NextYCommand
			if lastSent.Name == "loadSavedData" {
				var params *UserParams
				params , _ = readUser(remoteClient)
				remoteClient.Params = params
				bookingCenterId = remoteClient.Params.BookingPrefs.CenterID
			}
		} else if (userInput == "n" || userInput == "no") && lastSent.NextNCommand != "" {
			nextCmd = lastSent.NextNCommand
			if lastSent.Name == "loadSavedData" {
				resetParams(remoteClient)
				writeUser(remoteClient)
			}
		}
		if lastSent.Name == "search" && bookingCenterId > 0 {
			nextCmd = "search"
		} else if lastSent.Name == "otp" && otpTransactionId != "" {
			bearerToken, err = confirmOTP(userInput, remoteClient.Params.OTPTxnDetails.TXNId)
			if err == nil && bearerToken != "" {
				remoteClient.Params.OTPTxnDetails.BearerToken = bearerToken
				writeUser(remoteClient)
			}
		} else if lastSent.Name == "readCAPTCHA" {
			nextCmd = "bookingConfirmation"
		}
		lastSent.NextCommand = nextCmd
		sendResponse(remoteClient, nextCmd)
	}
}

func saveUserInput(remoteClient *RemoteClient) {
	var userInput = remoteClient.Received.Text
	var lastSent = remoteClient.LastSent
	if lastSent.Name == "vaccine" {
		remoteClient.Params.State = userInput
		writeUser(remoteClient)
	} else if lastSent.Name == "district" {
		remoteClient.Params.District = userInput
		writeUser(remoteClient)
	} else if lastSent.Name == "age" {
		if age, err := strconv.Atoi(userInput); err == nil {
			remoteClient.Params.Age = age
			writeUser(remoteClient)
		}
	} else if lastSent.Name == "otp" {
		if otp, err := strconv.Atoi(userInput); err == nil {
			remoteClient.Params.OTP = otp
			writeUser(remoteClient)
		}
	} else if lastSent.Name == "search" {
		var err error
		if bookingCenterId, err = strconv.Atoi(userInput); err == nil {
			remoteClient.Params.BookingPrefs.CenterID = bookingCenterId
			writeUser(remoteClient)
		}
	} else if lastSent.Name == "readCAPTCHA" {
		remoteClient.Params.CAPTCHA = userInput
		writeUser(remoteClient)
	}
}

func resetParams(remoteClient *RemoteClient) {
	bookingCenterId = 0
	stateID = 0
	districtID = 0
	remoteClient.Params = &UserParams{}
	remoteClient.Params.CAPTCHA = ""
	remoteClient.Params.OTPTxnDetails = &OTPTxn{}
	remoteClient.Params.BookingPrefs = &BookingSlot{}
	remoteClient.Params.Beneficiaries = &BeneficiaryList{}
}

func updateClient(remoteClient *RemoteClient, cmd *Command) {
	remoteClient.LastSent = cmd
	remoteClient.LastReceived = remoteClient.Received
	remoteClient.Received = pkWhatsApp.Message{}
	fmt.Fprintf(os.Stderr, "LastSent: \n{Name: %s, NextCmd: %s}\n", cmd.Name, cmd.NextCommand)
}
