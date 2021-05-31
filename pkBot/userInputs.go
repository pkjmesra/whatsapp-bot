package pkBot
import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

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
