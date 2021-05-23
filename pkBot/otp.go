package pkBot

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"

	"github.com/pkg/errors"
)

type OTPTxn struct {
	TXNId 		string `json:"txnId"`
	BearerToken string
	// Message string `json:"message"`
}

type OTPConfirmTxn struct {
	TOKEN string `json:"token"`
}

var (
	otpTransactionId, bearerToken, lastOTP string
)

const (
	// This client secret is a ciphertext generated by using the AES key and then calling
	// aes.encrypt(CLIENT_USERNAME, AES_KEY) for CBC mode
	// and getting the string form of the result. See setWithDynamicKey method in
	// https://selfregistration.cowin.gov.in/main-es2015.a738dfab7b730e7c14ac.js:formatted
	CLIENT_SECRET	= "U2FsdGVkX1+OsCQvQsCompcYGsHbguKSzKrIAh4xUB/us+Z7ZTqUiCsbkR+fKtfft6wsmYmTYyGlzVqi3L/x/g=="
	// See constructor of one of the classes in
	// https://selfregistration.cowin.gov.in/main-es2015.a738dfab7b730e7c14ac.js:formatted
	AES_KEY 		= "CoWIN@$#&*(!@%^&"
	// See anonymouslogin method in 
	// https://selfregistration.cowin.gov.in/main-es2015.a738dfab7b730e7c14ac.js:formatted
	// If it's different for your browser client code, you may just put the breakpoints
	// for the methods listed above and get the client secret yourself or use any AES lib
	// to calculate your secret using the key given above and the username plaintext
	CLIENT_USERNAME	= "b5cab167-7977-4df1-8027-a63aa144f04e"
)

func generateOTP(mobile string, public bool) (*OTPTxn, error) {
	fmt.Println("Generating OTP for "+ mobile)
	txn := OTPTxn{}
	if len(mobile) <= 0 {
		return &txn, errors.New("Invalid mobile number.")
	}
	log.Print("Generating OTP for: ", mobile)
	urlFormat := generateMobileOTPURLFormat
	postBody := map[string]interface{}{"mobile": mobile[len(mobile)-10:], "secret": CLIENT_SECRET}
	if public {
		urlFormat = generateOTPPublicURLFormat
		postBody = map[string]interface{}{"mobile": mobile[len(mobile)-10:]}
	}
	bodyBytes, err := queryServer(urlFormat, "POST", postBody)
	if err = json.Unmarshal(bodyBytes, &txn); err != nil {
		log.Printf("Error generating OTP: %s", err.Error())
		return &txn, err
	}
	otpTransactionId = txn.TXNId
	log.Print("OTP TxnId: ", otpTransactionId)
	return &txn, err
}

func confirmOTP(otp string, txnId string) (string, error) {
	h := sha256.New()
	h.Write([]byte(otp))
	sum := h.Sum(nil)
	sha := hex.EncodeToString(sum)

	log.Printf("Confirming OTP for TxnId: %s, SHA: %s", otpTransactionId, sha)
	postBody := map[string]interface{}{"otp" : sha, "txnId": otpTransactionId}
	bodyBytes, err := queryServer(validateMobileOtpURLFormat, "POST", postBody)
	txn := OTPConfirmTxn{}
	if err = json.Unmarshal(bodyBytes, &txn); err != nil {
		log.Print(string(bodyBytes))
		return "", err
	}
	bearerToken = txn.TOKEN
	// sendwhatsapptext(bearerToken)
	log.Print("Bearer Token: ", bearerToken)
	otpTransactionId = ""
	lastOTP = ""
	return bearerToken, nil
}