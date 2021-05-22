package pkBot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"text/tabwriter"

	// "github.com/pkg/errors"
)

type BeneficiaryList struct {
	Beneficiaries []struct {
		ReferenceID    	string `json:"beneficiary_reference_id"`
		Name  			string `json:"name"`
		Birth 			string `json:"birth_year"`
		Gender  		string `json:"gender"`
		Mobile  		string `json:"mobile_number"`
		PhotoIdType  	string `json:"photo_id_type"`
		PhotoIdNumber  	string `json:"photo_id_number"`
		ComorbidityInd  string `json:"comorbidity_ind"`
		VaccinationStat string `json:"vaccination_status"`
		Vaccine  		string `json:"vaccine"`
		Dose1Date  		string `json:"dose1_date"`
		Dose2Date  		string `json:"dose2_date"`
		Appointments []struct {
			AppointmentID    	string `json:"appointment_id"`
			CenterID  			int    `json:"center_id"`
			Name 				string `json:"name"`
			StateName  			string `json:"state_name"`
			DistrictName  		string `json:"district_name"`
			BlockName  			string `json:"block_name"`
			From  				string `json:"from"`
			To  				string `json:"to"`
			Dose  				int    `json:"dose"`
			SessionID  			string `json:"session_id"`
			Date  				string `json:"date"`
			Slot  				string `json:"slot"`
		} `json:"appointments"`
	} `json:"beneficiaries"`
}

type AppointmentList struct {
	DoseAppointments []struct {
		
	} `json:"appointments"`
}

func getBeneficiaries() (*BeneficiaryList, error) {
	response, err := queryServer(beneficiariesURLFormat, "GET", nil)
	beneficiaryList := BeneficiaryList{}
	err = json.Unmarshal(response, &beneficiaryList)
	if err != nil {
		log.Printf("Error parsing response: %v",err.Error())
		return &beneficiaryList, err
	}
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 1, 8, 1, '\t', 0)

	for _, beneficiary := range beneficiaryList.Beneficiaries {
		fmt.Fprintln(w, fmt.Sprintf("\nBeneficiaryID\t  %s", beneficiary.ReferenceID))
		fmt.Fprintln(w, fmt.Sprintf("Name\t  %s", beneficiary.Name))
		fmt.Fprintln(w, fmt.Sprintf("PhotoIdType\t  %s", beneficiary.PhotoIdType))
		fmt.Fprintln(w, fmt.Sprintf("PhotoIdNumber\t  %s", beneficiary.PhotoIdNumber))
		fmt.Fprintln(w, fmt.Sprintf("Status\t  %s", beneficiary.VaccinationStat))
		fmt.Fprintln(w, fmt.Sprintf("Dose1Date\t  %s", beneficiary.Dose1Date))
		fmt.Fprintln(w, fmt.Sprintf("Dose2Date\t  %s", beneficiary.Dose2Date))
		for _, appt := range beneficiary.Appointments {
			fmt.Fprintln(w, fmt.Sprintf("AppointmentID\t  %s", appt.AppointmentID))
			fmt.Fprintln(w, fmt.Sprintf("SessionID\t  %s", appt.SessionID))
			fmt.Fprintln(w, fmt.Sprintf("Dose\t  %d", appt.Dose))
			fmt.Fprintln(w, fmt.Sprintf("Date\t  %s", appt.Date))
			fmt.Fprintln(w, fmt.Sprintf("Slot\t  %s\n", appt.Slot))
		}
	}
	
	if err := w.Flush(); err != nil {
		log.Printf("Error writing to buffer: %v",err.Error())
		return &beneficiaryList, err
	}
	if buf.Len() == 0 {
		return &beneficiaryList, nil
	}
	// sendwhatsapptext("Beneficiaries registered:\n" + buf.String())
	return &beneficiaryList, err
}
