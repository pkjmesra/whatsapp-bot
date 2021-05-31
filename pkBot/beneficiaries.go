package pkBot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"text/tabwriter"

	// "github.com/pkg/errors"
)
var (beneficiariesList *BeneficiaryList)

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
	Description 		string
	EligibleCount		int
}

type AppointmentList struct {
	DoseAppointments []struct {
		
	} `json:"appointments"`
}

func getBeneficiaries() (*BeneficiaryList, error) {
	response, err := queryServer(beneficiariesURLFormat, "GET", nil)
	beneficiaryList := BeneficiaryList{}
	beneficiaryList.Description = ""
	err = json.Unmarshal(response, &beneficiaryList)
	if err != nil {
		log.Printf("Error parsing response: %v",err.Error())
		return &beneficiaryList, err
	}
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 1, 8, 1, '\t', 0)
	count := 0
	for _, beneficiary := range beneficiaryList.Beneficiaries {
		count++
		fmt.Fprintln(w, fmt.Sprintf("\n\n*( %d )*", count))
		fmt.Fprintln(w, fmt.Sprintf("BeneficiaryID : *%s*", beneficiary.ReferenceID))
		fmt.Fprintln(w, fmt.Sprintf("Name          : *%s*", beneficiary.Name))
		fmt.Fprintln(w, fmt.Sprintf("PhotoIdType   : %s", beneficiary.PhotoIdType))
		fmt.Fprintln(w, fmt.Sprintf("PhotoIdNumber : %s", beneficiary.PhotoIdNumber))
		fmt.Fprintln(w, fmt.Sprintf("Status        : %s", beneficiary.VaccinationStat))
		fmt.Fprintln(w, fmt.Sprintf("Dose1Date     : %s", beneficiary.Dose1Date))
		fmt.Fprintln(w, fmt.Sprintf("Dose2Date     : %s", beneficiary.Dose2Date))
		for _, appt := range beneficiary.Appointments {
			fmt.Fprintln(w, fmt.Sprintf("AppointmentID : %s", appt.AppointmentID))
			fmt.Fprintln(w, fmt.Sprintf("SessionID     : %s", appt.SessionID))
			fmt.Fprintln(w, fmt.Sprintf("Dose:         : %d", appt.Dose))
			fmt.Fprintln(w, fmt.Sprintf("Date:         : %s", appt.Date))
			fmt.Fprintln(w, fmt.Sprintf("Slot          : %s\n", appt.Slot))
		}
	}
	
	if err := w.Flush(); err != nil {
		log.Printf("Error writing to buffer: %v",err.Error())
		return &beneficiaryList, err
	}
	if buf.Len() == 0 {
		return &beneficiaryList, nil
	} else {
		beneficiaryList.Description = buf.String()
	}
	return &beneficiaryList, err
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
