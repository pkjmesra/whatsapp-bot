package pkBot

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

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
	PollingInterval 	int
	PollWaitCounter 	int
}

type UserParams struct {
	State 			string
	District		string
	StateID 		int
	DistrictID		int
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
	// if message.Info.FromMe || remoteClient.RemoteJID == remoteClient.RemoteMobileNumber + "@s.whatsapp.net" {
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
	// } else {
	// 	// remoteClient.Host.SendText(remoteClient.RemoteJID, "Hello from *github*!")
	// }
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

func downloadcertificate(remoteClient *RemoteClient, cmd *Command) {
	fmt.Println(remoteClient.RemoteJID + ":Now downloading certificate")
	response, err := queryServer(certificateURLFormat, "GET", nil)
	if err != nil {
		fmt.Println(remoteClient.RemoteJID + ":Restarting because failed getting certificate for " + remoteClient.RemoteMobileNumber)
		remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ErrorResponse2)
		return
	} else {
		certFile, err := os.Create(os.TempDir() + "Anubhav.pdf")
		if err != nil {
		 	fmt.Println(err)
		}
		defer certFile.Close()
		_ , err = certFile.Write(response)
		if err != nil {
		 	fmt.Println(err)
		}
		fmt.Println(remoteClient.RemoteJID + ":Certificate data saved into :" + os.TempDir() + "Anubhav.pdf")
	}
}

func handleAdhocBookingRequest(remoteClient *RemoteClient, cmd *Command) {
	fmt.Println(remoteClient.RemoteJID + ":Received booking request for pre-configured data")
	params := remoteClient.Params
	if params.State != "" && params.District != "" && params.Age > 0 {
		fmt.Println(remoteClient.RemoteJID + ":Pre-configured data found")
		if remoteClient.LastSent.NextCommand == "" {
			cd := Command{NextCommand:"search"}
			remoteClient.LastSent = &cd
		}
		updateClient(remoteClient, cmd)
		SendResponse(remoteClient, cmd.NextCommand)
		return
	} else {
		fmt.Println(remoteClient.RemoteJID + ":Restarting because pre-configured data not found")
		SendResponse(remoteClient, "")
	}
	updateClient(remoteClient, cmd)
	return
}

func handleSearchRequest(remoteClient *RemoteClient, cmd *Command) {
	bk, err := Search(remoteClient)
	if err != nil{
		fmt.Println(remoteClient.RemoteJID + ":Restarting because error while searching for slots")
		SendResponse(remoteClient, "")
	} else {
		if bk.Description == "" {
			fmt.Println(remoteClient.RemoteJID + ":Restarting because no slot found from the search")
			cmd.ToBeSent = cmd.ErrorResponse1
			remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ToBeSent)
			SendResponse(remoteClient, "")
			return
		}
		if bookingCenterId <= 0 && !remoteClient.Params.BookingPrefs.BookAnySlot {
			fmt.Println(remoteClient.RemoteJID + ":All slots being shared with user since preferred booking center is not set")
			remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ToBeSent + bk.Description)
		}
		if (bookingCenterId > 0 || remoteClient.Params.BookingPrefs.BookAnySlot) && bk.Description != "" {
			fmt.Println(remoteClient.RemoteJID + ":Preferred booking center is set already and center has slots")
			bk.BookAnySlot = remoteClient.Params.BookingPrefs.BookAnySlot
			remoteClient.Params.BookingPrefs = bk
			writeUser(remoteClient)
			askUserForOTP(remoteClient, cmd)
		}
		updateClient(remoteClient, cmd)
	}
}

func handleUserInputs(remoteClient *RemoteClient, cmd *Command) {
	if cmd.Name == "bookanyslot" {
		cmd.ToBeSent = fmt.Sprintf(cmd.ToBeSent, remoteClient.Params.District)
	} else if cmd.Name == "otp" {
		fmt.Println(remoteClient.RemoteJID + ":Generated OTP for booking. Now sending text to the user to receive OTP.")
		br := remoteClient.Params.BookingPrefs
		var centerName string
		if len(br.PotentialSessions) > 0 {
			ps := br.PotentialSessions[0]
			centerName = ps.CenterName
			if len(br.PotentialSessions) > 1 {
				centerName = centerName + " and others"
			}
		}
		cmd.ToBeSent = fmt.Sprintf(cmd.ToBeSent, len(br.PotentialSessions), centerName, br.TotalDose1Available)
	}
	remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ToBeSent)
	updateClient(remoteClient, cmd)
}

func handleBookingRequest(remoteClient *RemoteClient, cmd *Command) {
	var cnf string
	var err error
	cnf, err = bookAppointment(remoteClient.Params.Beneficiaries, remoteClient.Params.CAPTCHA, remoteClient.Params.BookingPrefs)
	if err == nil && cnf != "" {
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
		fmt.Println(remoteClient.RemoteJID + ":Restarting because appointment confirmation failed.")
		SendResponse(remoteClient, "")
	}
	resetPollingInterval(remoteClient, globalPollingInterval)
}

func askUserForOTP(remoteClient *RemoteClient, cmd *Command) {
	var txn *OTPTxn
	var err error
	resetPollingInterval(remoteClient, defaultSearchInterval)
	fmt.Println(remoteClient.RemoteJID + ":Generating OTP for booking now")
	txn, err = generateOTP(remoteClient.RemoteMobileNumber, false)
	if err == nil && txn.TXNId != "" {
		remoteClient.Params.OTPTxnDetails = txn
		writeUser(remoteClient)
		updateClient(remoteClient, cmd)
		SendResponse(remoteClient, cmd.NextCommand)
		return
	} else {
		fmt.Println(remoteClient.RemoteJID + ":Restarting because failed generating OTP for " + remoteClient.RemoteMobileNumber)
		remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ErrorResponse2)
		SendResponse(remoteClient, "")
		return
	}
}

func generateOTPForBearerToken(remoteClient *RemoteClient, cmd *Command) {
	var txn *OTPTxn
	var err error
	txn, err = generateOTP(remoteClient.RemoteMobileNumber, false)
	if err == nil && txn.TXNId != "" {
		remoteClient.Params.OTPTxnDetails = txn
		writeUser(remoteClient)
		remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ToBeSent)
		updateClient(remoteClient, cmd)
		// SendResponse(remoteClient, cmd.NextCommand)
	} else {
		fmt.Println(remoteClient.RemoteJID + ":Restarting because failed generating OTP for " + remoteClient.RemoteMobileNumber)
		remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ErrorResponse2)
		SendResponse(remoteClient, "")
	}
}

func Search(remoteClient *RemoteClient) (*BookingSlot, error) {
	remoteClient.Params, _ = readUser(remoteClient)
	bk, err := searchByStateDistrict(remoteClient)
	if err != nil {
		fmt.Fprintf(os.Stderr,remoteClient.RemoteJID + ":(Search)Error while searching for slots:%v\n", err)
	}
	return bk, err
}

func getBeneficiariesForRemoteUser(remoteClient *RemoteClient, cmd *Command, force bool){
	var beneficiariesList *BeneficiaryList
	var err error
	savedBeneficiaries := remoteClient.Params.Beneficiaries
	if force || len(savedBeneficiaries.Beneficiaries) == 0 || savedBeneficiaries.EligibleCount == 0 {
		beneficiariesList, err = getBeneficiaries()
		if err == nil && beneficiariesList.Description != ""{
			eligible, _ := eligibleBeneficiaries(beneficiariesList)
			beneficiariesList.EligibleCount = len(eligible)
			remoteClient.Params.Beneficiaries = beneficiariesList
			remoteClient.Params.BookingPrefs.EligibleCount = len(eligible)
			writeUser(remoteClient)
			cmd.ToBeSent = fmt.Sprintf(cmd.ToBeSent, beneficiariesList.Description)
			remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ToBeSent)
		}
	}
	updateClient(remoteClient, cmd)
	SendResponse(remoteClient, cmd.NextCommand)
}

func sendCAPTCHA(remoteClient *RemoteClient, cmd *Command) {
	fmt.Println(remoteClient.RemoteJID + ":Now sending CAPTCHA request")
	resetPollingInterval(remoteClient, defaultSearchInterval)
	fileName := remoteClient.RemoteMobileNumber + "_captcha"
	err := getCaptchaSVG(fileName + ".svg")
	var f string
	if err != nil {
		fmt.Println(remoteClient.RemoteJID + ":Restarting because failed getting CAPTCHA(svg) for " + remoteClient.RemoteMobileNumber)
		remoteClient.Host.SendText(remoteClient.RemoteJID, fmt.Sprintf(cmd.ErrorResponse1, remoteClient.RemoteMobileNumber))
		SendResponse(remoteClient, "")
		return
	}
	err = exportToPng(fileName + ".svg", fileName + ".png")
	if err != nil {
		fmt.Println(remoteClient.RemoteJID + ":Restarting because Failed getting CAPTCHA(png) for " + remoteClient.RemoteMobileNumber)
		remoteClient.Host.SendText(remoteClient.RemoteJID, fmt.Sprintf(cmd.ErrorResponse1, remoteClient.RemoteMobileNumber))
		SendResponse(remoteClient, "")
		return
	}
	_ , f, _ , err = getPngImage(fileName + ".png")
	if err != nil {
		fmt.Println(remoteClient.RemoteJID + ":Restarting because failed getting CAPTCHA(getpng) for " + remoteClient.RemoteMobileNumber)
		remoteClient.Host.SendText(remoteClient.RemoteJID, fmt.Sprintf(cmd.ErrorResponse1, remoteClient.RemoteMobileNumber))
		SendResponse(remoteClient, "")
		return
	}
	f, err = getJPEGImage(fileName + ".png")
	if err != nil {
		fmt.Println(remoteClient.RemoteJID + ":Restarting because failed getting CAPTCHA(jpg) for " + remoteClient.RemoteMobileNumber)
		remoteClient.Host.SendText(remoteClient.RemoteJID, fmt.Sprintf(cmd.ErrorResponse1, remoteClient.RemoteMobileNumber))
		SendResponse(remoteClient, "")
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
			SendResponse(remoteClient, lastSent.NextCommand)
		} else {
			fmt.Println(remoteClient.RemoteJID + ":Restarting because ExpectedResponse is different from userInput")
			SendResponse(remoteClient, "")
		}
	} else {
		userInput = strings.ToLower(userInput)
		nextCmd := lastSent.NextCommand
		if (userInput == "y" || userInput == "yes") && lastSent.NextYCommand != "" {
			nextCmd = lastSent.NextYCommand
			var params *UserParams
			params , _ = readUser(remoteClient)
			remoteClient.Params = params
			if lastSent.Name == "loadsaveddata" {
				bookingCenterId = remoteClient.Params.BookingPrefs.CenterID
				fmt.Println(remoteClient.RemoteJID + ":Saved  bookingCenterId")
			} else if lastSent.Name == "bookanyslot" {
				remoteClient.Params.BookingPrefs.BookAnySlot = true
				writeUser(remoteClient)
				fmt.Fprintf(os.Stderr, remoteClient.RemoteJID + ":Saved BookAnySlot. Set to %t.", remoteClient.Params.BookingPrefs.BookAnySlot)
			}
		} else if (userInput == "n" || userInput == "no") && lastSent.NextNCommand != "" {
			nextCmd = lastSent.NextNCommand
			var params *UserParams
			params , _ = readUser(remoteClient)
			remoteClient.Params = params
			if lastSent.Name == "loadsaveddata" {
				resetParams(remoteClient, false)
			} else if lastSent.Name == "bookanyslot" {
				remoteClient.Params.BookingPrefs.BookAnySlot = false
				writeUser(remoteClient)
				fmt.Println(remoteClient.RemoteJID + ":Saved  BookAnySlot. Set to false.")
			}
		}
		if lastSent.Name == "search" && bookingCenterId > 0 && !remoteClient.Params.BookingPrefs.BookAnySlot {
			nextCmd = "search"
		} else if (lastSent.Name == "otp" || lastSent.Name == "otpbeneficiary" || lastSent.Name == "certificate") && otpTransactionId != "" {
			bearerToken, err = confirmOTP(userInput, remoteClient.Params.OTPTxnDetails.TXNId)
			if err == nil && bearerToken != "" {
				fmt.Println(remoteClient.RemoteJID + ":BearerToken received:" + bearerToken)
				remoteClient.Params.OTPTxnDetails.BearerToken = bearerToken
				writeUser(remoteClient)
			}
		} else if lastSent.Name == "readcaptcha" {
			nextCmd = "bookingconfirmation"
		}
		lastSent.NextCommand = nextCmd
		SendResponse(remoteClient, nextCmd)
	}
}

func saveUserInput(remoteClient *RemoteClient) {
	var userInput = remoteClient.Received.Text
	var lastSent = remoteClient.LastSent
	if lastSent.Name == "vaccine" || lastSent.Name == "state" {
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
	} else if lastSent.Name == "otp" || lastSent.Name == "otpbeneficiary" || lastSent.Name == "certificate" {
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
	} else if lastSent.Name == "readcaptcha" {
		remoteClient.Params.CAPTCHA = userInput
		writeUser(remoteClient)
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

