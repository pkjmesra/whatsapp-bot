package pkBot

import (
	"encoding/json"
	// "log"
	"fmt"
	"os"
	"strings"
	"github.com/pkg/errors"
)

type AppointmentConfirmation struct {
	ConfirmationId string `json:"appointment_confirmation_no"`
	// Message string `json:"message"`
}

type ApptRequestError struct {
	Error    	string `json:"error"`
	ErrorCode  	string `json:"errorCode"`
}

func bookAppointment(beneficiaryList *BeneficiaryList, captcha string, bookingSlot *BookingSlot) (string, error) {
	fmt.Fprintf(os.Stderr, "Booking will be attempted for potentials sessions: %d", len(bookingSlot.PotentialSessions))
	beneficiaries, beneficiaryId := eligibleBeneficiaries(beneficiaryList)
	var cnf string
	err := errors.New("Booking failed because no request could be made (None of the available centers had as many doses left)")

	var potentialSession PotentialSession
	// postBody := map[string]interface{}{"center_id": , "dose": 1, "captcha": captcha,"session_id": bookingSlot.SessionID, "slot": bookingSlot.Slot, "beneficiaries": beneficiaries}
	if bookingSlot.BookAnySlot {
		for _, potentialSession = range bookingSlot.PotentialSessions {
			if int(potentialSession.Dose1Capacity) >= len(beneficiaries) {
				c := potentialSession.CenterID
				s := potentialSession.SessionID
				sl := potentialSession.Slots[0]
				fmt.Fprintf(os.Stderr, "Booking an appt with centerid:%d, captcha:%s, session:%s, slot:%s, and beneficiaries:%s\n", c, captcha, s, sl, beneficiaryId)
				cnf, err = tryBookAppointment(c,captcha, s, sl, beneficiaries)
				if err == nil && cnf != ""{
					cnf = cnf + "\nCenter:" + potentialSession.CenterName + "\nAddress:" + potentialSession.CenterAddress + "\nSlot:" + sl + "\n"
				}
				break
			}
		}
	}
	return cnf, err
}

func tryBookAppointment(centerId int, captcha, sessionId, slot string, beneficiaries []string) (string, error) {
	postBody := map[string]interface{}{"center_id": centerId, "dose": 1, "captcha": captcha,"session_id": sessionId, "slot": slot, "beneficiaries": beneficiaries}
	bodyBytes, err := queryServer(scheduleURLFormat, "POST", postBody)
	cnf := AppointmentConfirmation{}
	if err = json.Unmarshal(bodyBytes, &cnf); err != nil {
		fmt.Println("Error in booking!")
	}
	if cnf.ConfirmationId == "" || err != nil {
		aptErr := ApptRequestError{}
		if err = json.Unmarshal(bodyBytes, &aptErr); err != nil {
			fmt.Println("Error scheduling: " + err.Error())
		}
		fmt.Fprintf(os.Stderr, "ErrorCode:%s , Error:%s\n", aptErr.ErrorCode, aptErr.Error)
		return "", errors.New(fmt.Sprintf("Booking failed with statusCode: *%s : %s*\n", aptErr.ErrorCode, aptErr.Error))
	}
	fmt.Println("AppointmentID confirmed with confirmation ID:" + cnf.ConfirmationId)
	return cnf.ConfirmationId, nil
}

func eligibleBeneficiaries(beneficiaryList *BeneficiaryList) ([]string, string) {
	beneficiaryId := ""
	for _, beneficiary := range beneficiaryList.Beneficiaries {
		if beneficiary.VaccinationStat == "Not Vaccinated" {
			if beneficiaryId == "" {
				beneficiaryId = beneficiary.ReferenceID
			} else {
				beneficiaryId = beneficiaryId + "," + beneficiary.ReferenceID
			}
		}
	}
	beneficiaries:= strings.Split(beneficiaryId, ",") //make([]string,1)
	fmt.Fprintf(os.Stderr, "Booking will be attempted for these %d beneficiaries: %s", len(beneficiaries),beneficiaryId)
	return beneficiaries, beneficiaryId
}

// func sortSessions(bookingSlot *BookingSlot) {
// 	sort.Slice(bookingSlot.PotentialSessions, func(s1, s2 PotentialSession) bool {
// 	    return s1.Dose1Capacity > s2.Dose1Capacity
// 	})
// 	if bookingSlot.BookAnySlot {
// 		for _, potentialSession = range bookingSlot.PotentialSessions {
// 			if int(potentialSession.Dose1Capacity) >= len(beneficiaries) {
// 				c := potentialSession.CenterID
// 				s := potentialSession.SessionID
// 				sl := potentialSession.Slots[0]
// 				fmt.Fprintf(os.Stderr, "Booking an appt with centerid:%d, captcha:%s, session:%s, slot:%s, and beneficiaries:%s\n", c, captcha, s, sl, beneficiaryId)
// 				cnf, err = tryBookAppointment(c,captcha, s, sl, beneficiaries)
// 				if err == nil && cnf != ""{
// 					cnf = cnf + "\nCenter:" + potentialSession.CenterName + "\nAddress:" + potentialSession.CenterAddress + "\nSlot:" + sl + "\n"
// 				}
// 				break
// 			}
// 		}
// 	}
// }

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
