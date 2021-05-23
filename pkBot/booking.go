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
	//beneficiaries[0] = beneficiaryId
	postBody := map[string]interface{}{"center_id": bookingSlot.CenterID, "dose": 1, "captcha": captcha,"session_id": bookingSlot.SessionID, "slot": bookingSlot.Slot, "beneficiaries": beneficiaries}
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
