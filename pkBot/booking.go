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
	beneficiaryId := ""
	fmt.Fprintf(os.Stderr, "Booking will be attempted for potentials sessions: %d", len(bookingSlot.PotentialSessions))
	// outer:
	for _, beneficiary := range beneficiaryList.Beneficiaries {
		if beneficiary.VaccinationStat == "Not Vaccinated" {
			if beneficiaryId == "" {
				beneficiaryId = beneficiary.ReferenceID
			} else {
				beneficiaryId = beneficiaryId + "," + beneficiary.ReferenceID
			}
			// break outer
		}
	}
	beneficiaries:= strings.Split(beneficiaryId, ",") //make([]string,1)
	var cnf string
	err := errors.New("Booking failed because no request could be made")

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
	fmt.Println("AppointmentID confirmed:" + cnf.ConfirmationId)
	return cnf.ConfirmationId, nil
}