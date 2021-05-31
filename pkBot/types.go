package pkBot

import (
	"github.com/pkjmesra/whatsapp-bot/pkWhatsApp"
)

type PotentialSession struct {
	CenterID 		  int
	CenterName 		  string
	CenterAddress 	  string
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
		EligibleCount int
		TotalDose1Available int
}

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

type Subscriptions struct {
		Subscribers 	[]Subscriber
		NonSubscribers 	[]Subscriber
}

type Subscriber struct {
		RemoteJID 		string
		MobileNumber 	string
		Date 			string
}

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
