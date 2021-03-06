package pkBot

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
	"text/tabwriter"
	"time"
	"os"

	"github.com/pkg/errors"
)

var (
	slotsAvailable bool
	bookingSlot *BookingSlot
	pinCode, bookingCenterId int
)

func timeNow() string {
	return time.Now().Format("02-01-2006")
}

func searchByPincode(remoteClient *RemoteClient, pinCode string) (*BookingSlot, error) {
	bk := remoteClient.Params.BookingPrefs
	response, err := queryServer(fmt.Sprintf(calendarByPinURLFormat, pinCode, timeNow()), "GET", nil)
	if err != nil {
		return bk, errors.Wrap(err, "Failed to fetch appointment sessions")
	}
	return getAvailableSessions(response, remoteClient.Params.Age, pinCode, bk)
}

func getStateIDByName(state string) (int, error) {
	response, err := queryServer(listStatesURLFormat, "GET", nil)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to list states")
	}
	states := StateList{}
	if err := json.Unmarshal(response, &states); err != nil {
		return 0, err
	}
	for _, s := range states.States {
		if strings.ToLower(s.StateName) == strings.ToLower(state) {
			// log.Printf("State Details - ID: %d, Name: %s", s.StateID, s.StateName)
			return s.StateID, nil
		}
	}
	return 0, errors.New("Invalid state name passed")
}

func getDistrictIDByName(stateID int, district string) (int, error) {
	response, err := queryServer(fmt.Sprintf(listDistrictsURLFormat, stateID), "GET", nil)
	if err != nil {
		return 0, errors.Wrap(err, "Failed to list states")
	}
	dl := DistrictList{}
	if err := json.Unmarshal(response, &dl); err != nil {
		return 0, err
	}
	for _, d := range dl.Districts {
		if strings.ToLower(d.DistrictName) == strings.ToLower(district) {
			// log.Printf("District Details - ID: %d, Name: %s", d.DistrictID, d.DistrictName)
			return d.DistrictID, nil
		}
	}
	return 0, errors.New("Invalid district name passed")
}

func searchByStateDistrict(remoteClient *RemoteClient) (*BookingSlot, error) {
	var err1 error
	params := remoteClient.Params
	bk := params.BookingPrefs
	if params.StateID == 0 {
		params.StateID, err1 = getStateIDByName(params.State)
		if err1 != nil {
			return bk, err1
		}
	}
	if params.DistrictID == 0 {
		params.DistrictID, err1 = getDistrictIDByName(params.StateID, params.District)
		if err1 != nil {
			return bk, err1
		}
	}
	fmt.Fprintf(os.Stderr, remoteClient.RemoteJID + ":Searching with age:%d, state:%s, district:%s\n", params.Age, params.State, params.District)
	response, err := queryServer(fmt.Sprintf(calendarByDistrictURLFormat, params.DistrictID, timeNow()), "GET", nil)
	if err != nil {
		return bk, errors.Wrap(err, "Failed to fetch appointment sessions")
	}
	var criteria = params.State + "," + params.District
	// var ck *BookingSlot
	bk, err = getAvailableSessions(response, params.Age, criteria, bk)
	return bk, err
}

func getAvailableSessions(response []byte, age int, criteria string, bk *BookingSlot) (*BookingSlot, error) {
	ps := []PotentialSession{}
	bk.PotentialSessions = ps // := BookingSlot{Available:false, Preferred:false, PotentialSessions: ps}
	if response == nil {
		// log.Printf("Received unexpected response, rechecking after %v seconds", interval)
		return bk, nil
	}
	appnts := Appointments{}
	err := json.Unmarshal(response, &appnts)
	if err != nil {
		return bk, err
	}
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 1, 8, 1, '\t', 0)

	count := 0
	bk.TotalDose1Available = 0
	outer:
	for _, center := range appnts.Centers {
		for _, s := range center.Sessions {
			capacity := s.Dose1Capacity
			if s.AvailableCapacity > 0 {
				fmt.Fprintf(os.Stderr, "CenterID: %d , AvailableCapacity:%.0f, Dose1Capacity (%d years):%.0f\n", center.CenterID, s.AvailableCapacity, s.MinAgeLimit, capacity)
			}
			if s.MinAgeLimit <= age && (int(capacity) >= bk.EligibleCount ) { //|| s.Dose2Capacity != 0) {
				if bk.CenterID > 0 {
					if bk.CenterID == center.CenterID {
						fmt.Fprintf(os.Stderr, "AvailableCapacity %.0f for selected center:%d\n", s.AvailableCapacity, center.CenterID)
						bk.Preferred = true
						bk.CenterID = center.CenterID
						bk.SessionID = s.SessionID
						bk.CenterName = center.Name
						bk.Slot = s.Slots[0]
						center.Name = "(Preferred) :" + center.Name
					}
				}
				if bk.BookAnySlot && capacity != 0 {
					fmt.Fprintf(os.Stderr, "CenterID: %d , AvailableCapacity:%.0f, Dose1Capacity:%.0f\n", center.CenterID, s.AvailableCapacity,capacity)
					s.CenterID = center.CenterID
					bk.TotalDose1Available = bk.TotalDose1Available + int(capacity)
					newSession := PotentialSession{CenterID: s.CenterID,
									CenterName : center.Name,
									CenterAddress : center.Address,
									SessionID: s.SessionID,
									Date:s.Date,
									AvailableCapacity: s.AvailableCapacity,
									MinAgeLimit: s.MinAgeLimit,
									Vaccine: s.Vaccine,
									Dose1Capacity:capacity,
									Dose2Capacity: s.Dose2Capacity,
									Slots: s.Slots}
					bk.PotentialSessions = append(bk.PotentialSessions, newSession)
					fmt.Fprintf(os.Stderr, "New Potential Session added. Total Count:%d\n", len(bk.PotentialSessions))
				}
				if (bk.Preferred || bk.CenterID <= 0 || bk.BookAnySlot) && capacity != 0 {
					count++
					fmt.Fprintln(w, fmt.Sprintf("*(%d). Center\t  %s, %s, %s, %s, %d*", count, center.Name, center.Address, center.DistrictName, center.StateName, center.Pincode))
					fmt.Fprintln(w, fmt.Sprintf("Fee\t  %s", center.FeeType))
					fmt.Fprintln(w, fmt.Sprintf("CenterID\t  %d", center.CenterID))
					if len(center.VaccineFees) != 0 {
						fmt.Fprintln(w, fmt.Sprintf("Vaccine\t"))
					}
					for _, v := range center.VaccineFees {
						fmt.Fprintln(w, fmt.Sprintf("\tName\t  %s", v.Vaccine))
						fmt.Fprintln(w, fmt.Sprintf("\tFees\t  %s", v.Fee))
					}
					// fmt.Fprintln(w, fmt.Sprintf("Sessions\t"))
					fmt.Fprintln(w, fmt.Sprintf("\t*Date\t  %s*", s.Date))
					fmt.Fprintln(w, fmt.Sprintf("\t*Dose1Capacity\t  %.0f*", capacity))
					fmt.Fprintln(w, fmt.Sprintf("\t*Dose2Capacity\t  %.0f*", s.Dose2Capacity))
					fmt.Fprintln(w, fmt.Sprintf("\tMinAgeLimit\t  %d", s.MinAgeLimit))
					fmt.Fprintln(w, fmt.Sprintf("\tVaccine\t  %s", s.Vaccine))
					fmt.Fprintln(w, fmt.Sprintf("\tSlots"))
					for _, slot := range s.Slots {
						fmt.Fprintln(w, fmt.Sprintf("\t\t  %s", slot))
					}
				}
				if bk.Preferred {
					break outer
				}
			}
		}
	}
	if err := w.Flush(); err != nil {
		return bk, err
	}
	if buf.Len() == 0 {
		if slotsAvailable {
			slotsAvailable = false
			bk.Available = slotsAvailable
		}
		bk.Description = buf.String()
		return bk, nil
	}
	slotsAvailable = true
	bk.Available = slotsAvailable
	bk.Description = buf.String()
	return bk, nil
}

func Search(remoteClient *RemoteClient) (*BookingSlot, error) {
	remoteClient.Params, _ = readUser(remoteClient)
	bk, err := searchByStateDistrict(remoteClient)
	if err != nil {
		fmt.Fprintf(os.Stderr,remoteClient.RemoteJID + ":(Search)Error while searching for slots:%v\n", err)
	}
	return bk, err
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
