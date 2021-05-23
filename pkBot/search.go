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

type PotentialSession struct {
	CenterID 		  int
	SessionID         string
	Date              string
	AvailableCapacity float64
	MinAgeLimit       int
	Vaccine           string
	Dose1Capacity     float64
	Dose2Capacity     float64
	Slots             []string
}
type BookingSlot struct {
		Available 	bool
		Preferred 	bool
		CenterID 	int
		BookAnySlot bool
		CenterName  string
		SessionID 	string
		Slot 		string
		Description string
		PotentialSessions []PotentialSession
	}

var (
	slotsAvailable bool
	bookingSlot *BookingSlot
	age, pinCode, bookingCenterId, stateID, districtID int
)

type StateList struct {
	States []struct {
		StateID    int    `json:"state_id"`
		StateName  string `json:"state_name"`
		StateNameL string `json:"state_name_l"`
	} `json:"states"`
	TTL int `json:"ttl"`
}

type DistrictList struct {
	Districts []struct {
		StateID       int    `json:"state_id"`
		DistrictID    int    `json:"district_id"`
		DistrictName  string `json:"district_name"`
		DistrictNameL string `json:"district_name_l"`
	} `json:"districts"`
	TTL int `json:"ttl"`
}

type Appointments struct {
	Centers []struct {
		CenterID      int     `json:"center_id"`
		Name          string  `json:"name"`
		NameL         string  `json:"name_l"`
		StateName     string  `json:"state_name"`
		StateNameL    string  `json:"state_name_l"`
		DistrictName  string  `json:"district_name"`
		DistrictNameL string  `json:"district_name_l"`
		BlockName     string  `json:"block_name"`
		BlockNameL    string  `json:"block_name_l"`
		Pincode       int     `json:"pincode"`
		Address       string  `json:"address"`
		Lat           float64 `json:"lat"`
		Long          float64 `json:"long"`
		From          string  `json:"from"`
		To            string  `json:"to"`
		FeeType       string  `json:"fee_type"`
		VaccineFees   []struct {
			Vaccine string `json:"vaccine"`
			Fee     string `json:"fee"`
		} `json:"vaccine_fees"`
		Sessions []struct {
			CenterID 		  int
			SessionID         string   `json:"session_id"`
			Date              string   `json:"date"`
			AvailableCapacity float64  `json:"available_capacity"`
			MinAgeLimit       int      `json:"min_age_limit"`
			Vaccine           string   `json:"vaccine"`
			Dose1Capacity     float64   `json:"available_capacity_dose1"`
			Dose2Capacity     float64   `json:"available_capacity_dose2"`
			Slots             []string `json:"slots"`
		} `json:"sessions"`
	} `json:"centers"`
}

func timeNow() string {
	return time.Now().Format("02-01-2006")
}

func searchByPincode(pinCode string) (*BookingSlot, error) {
	bk := BookingSlot{Available:false}
	response, err := queryServer(fmt.Sprintf(calendarByPinURLFormat, pinCode, timeNow()), "GET", nil)
	if err != nil {
		return &bk, errors.Wrap(err, "Failed to fetch appointment sessions")
	}
	return getAvailableSessions(response, age, pinCode, &bk)
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

func searchByStateDistrict(age int, state, district string, bk *BookingSlot) (*BookingSlot, error) {
	var err1 error
	// bk := BookingSlot{Available:false}
	if stateID == 0 {
		stateID, err1 = getStateIDByName(state)
		if err1 != nil {
			return bk, err1
		}
	}
	if districtID == 0 {
		districtID, err1 = getDistrictIDByName(stateID, district)
		if err1 != nil {
			return bk, err1
		}
	}
	fmt.Fprintf(os.Stderr, "Searching with age:%d, state:%s, district:%s\n", age, state, district)
	response, err := queryServer(fmt.Sprintf(calendarByDistrictURLFormat, districtID, timeNow()), "GET", nil)
	if err != nil {
		return bk, errors.Wrap(err, "Failed to fetch appointment sessions")
	}
	var criteria = state + "," + district
	// var ck *BookingSlot
	bk, err = getAvailableSessions(response, age, criteria, bk)
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
	outer:
	for _, center := range appnts.Centers {
		for _, s := range center.Sessions {
			capacity := s.Dose1Capacity
			// fmt.Fprintf(os.Stderr, "CenterID: %d , AvailableCapacity:%.0f\n", center.CenterID, s.AvailableCapacity)
			if s.MinAgeLimit <= age && (capacity != 0 ) { //|| s.Dose2Capacity != 0) {
				if bookingCenterId > 0 {
					if bookingCenterId == center.CenterID {
						// fmt.Fprintf(os.Stderr, "AvailableCapacity %.0f for selected center:%d\n", s.AvailableCapacity, center.CenterID)
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
					newSession := PotentialSession{CenterID: s.CenterID,
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
				if (bk.Preferred || bookingCenterId <= 0 || bk.BookAnySlot) && capacity != 0 {
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
