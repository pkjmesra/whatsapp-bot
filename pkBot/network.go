package pkBot

import (
	"bytes"
   	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	// "strings"
	"time"

	"github.com/pkg/errors"
)

// https://apisetu.gov.in/public/api/cowin
const (
	// See the constants in 
	// https://selfregistration.cowin.gov.in/main-es2015.a738dfab7b730e7c14ac.js:formatted
	baseURL                     = "https://cdn-api.co-vin.in/api"
	calendarByPinURLFormat      = "/v2/appointment/sessions/calendarByPin?pincode=%s&date=%s"
	calendarByDistrictURLFormat = "/v2/appointment/sessions/calendarByDistrict?district_id=%d&date=%s" //&vaccine=%s"
	listStatesURLFormat         = "/v2/admin/location/states"
	listDistrictsURLFormat      = "/v2/admin/location/districts/%d"

	generateOTPPublicURLFormat  = "/v2/auth/public/generateOTP"
	generateMobileOTPURLFormat  = "/v2/auth/generateMobileOTP"
	confirmOTPPublicURLFormat   = "/v2/auth/public/confirmOTP"
	validateMobileOtpURLFormat  = "/v2/auth/validateMobileOtp"
	beneficiariesURLFormat		= "/v2/appointment/beneficiaries"

	scheduleURLFormat			= "/v2/appointment/schedule"
	cancelURLFormat				= "/v2/appointment/cancel"

	captchaURLFormat			= "/v2/auth/getRecaptcha"

	certificateURLFormat 		= "/v2/registration/certificate/download?beneficiary_reference_id=%s"

	// admin_prefix_url: "https://cdn-api.co-vin.in/api/v1/admin"
	// admin_prefix_url_v2: "https://cdn-api.co-vin.in/api/v2/admin"
	// appointment_url: "https://cdn-api.co-vin.in/api/v1/appointment"
	// appointment_url_dir: "https://www.cowin.gov.in/api/v1/appointment"
	// appointment_url_v2: "https://cdn-api.co-vin.in/api/v2/appointment"
	// auth_prefix_url: "https://cdn-api.co-vin.in/api/v1/auth"
	// auth_prefix_v2: "https://cdn-api.co-vin.in/api/v2/auth"
	// beneficiary_registration_prefix_direct_url: "https://www.cowin.gov.in/api/v1/registration"
	// beneficiary_registration_prefix_url: "https://cdn-api.co-vin.in/api/v1/registration"
	// defaultLanguage: "en"
	// digilockerClientId: "E7D87B4B"
	// digilockerClientSecret: "dce34912c0a610dd0e32"
	// digilockerRedirectURI: "https://selfregistration.cowin.gov.in/selfregistration"
	// digilockerURI: "https://api.digitallocker.gov.in/public/oauth2/1/authorize?response_type=code&client_id=E7D87B4B&redirect_uri=https://selfregistration.cowin.gov.in/selfregistration&state={}"
	// rd_device_prefix_url: "https://localhost:11200"
	// recaptcha_site_key: "6LdNBNUZAAAAAAYITU9accPJNhMSzHMsuqc99RPE"
	// registration_url_v2: "https://www.cowin.gov.in/api/v2/registration"
	// registration_url_v2_cdn: "https://cdn-api.co-vin.in/api/v2/registration"
	// session_prefix_url: "https://cdn-api.co-vin.in/api/v1/session"
	// session_prefix_url_port: "https://cdn-api.co-vin.in:9005"

)

func queryServer(path string, method string, jsonBody map[string]interface{}) ([]byte, error) {
	var requestBody io.Reader
	if method == "POST" || jsonBody != nil {
		postBody, err := json.Marshal(jsonBody)
		if err != nil {
		    log.Fatal(err)
		}
	   	requestBody = bytes.NewBuffer(postBody)
	} else {
		requestBody = nil
	}
	req, err := http.NewRequest(method, baseURL+path, requestBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("authority", "cdn-api.co-vin.in")
	req.Header.Set("user-agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 11_2_3) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.90 Safari/537.36")
	req.Header.Set("accept", "application/json, text/plain, */*")
	req.Header.Set("accept-language", "en-GB,en-US;q=0.9,en;q=0.8")
	req.Header.Set("content-type", "application/json; charset=UTF-8")
	req.Header.Set("origin", "https://selfregistration.cowin.gov.in")
	req.Header.Set("referer", "https://selfregistration.cowin.gov.in/")
		

	if len(bearerToken) > 0 {
		req.Header.Set("authorization", "Bearer " + bearerToken)
	}
	// log.Print("Querying endpoint: ", baseURL+path)
	// Save a copy of this request for debugging.
	requestDump, err := httputil.DumpRequest(req, jsonBody != nil)
	if err != nil {
	  fmt.Println(err)
	  fmt.Println(string(requestDump))
	}

	var netClient = &http.Client{
	  Timeout: time.Second * 30,
	}

	resp, err := netClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()


	bodyBytes, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
		return nil, err
	}
	responseString := string(bodyBytes)
	if len(responseString) >= 2000 {
		log.Print("Response: ", responseString[:100])
	} else {
		log.Print("Response: ", responseString)
	}

	if resp.StatusCode != http.StatusOK {
		// Sometimes the API returns "Unauthenticated access!", do not fail in that case
		if resp.StatusCode == http.StatusUnauthorized {
			return bodyBytes, errors.New(fmt.Sprintf("Request failed with statusCode: %d", resp.StatusCode))
		}
		return bodyBytes, errors.New(fmt.Sprintf("Request failed with statusCode: %d", resp.StatusCode))
	}
	return bodyBytes, nil
}
