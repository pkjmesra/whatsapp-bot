package pkBot

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "path/filepath"
    "os"
    "strings"

)

type Command struct {
	Name      			string  `json:"name"`
	CommandType         string  `json:"commandType"`
	ToBeSent     		string  `json:"toBeSent"`
	ResponseType    	string  `json:"responseType"`
	ExpectedResponse    string  `json:"expectedResponse"`
	NextCommand  		string  `json:"nextCommand"`
	ErrorResponse1		string 	`json:"errorResponse1"`
	ErrorResponse2		string 	`json:"errorResponse2"`
	NextYCommand		string  `json:"nextYCommand"`
	NextNCommand		string  `json:"nextNCommand"`
}

var (
	root map[string]interface{}
	allSubscribersMap map[string]*Subscriber
	allSubscribers []Subscriber
)

func Initialize() *map[string]interface{} {
    jsonFile, err := os.Open("commands.json")
    if err != nil {
        fmt.Println(err)
    }
    defer jsonFile.Close()

    byteValue, _ := ioutil.ReadAll(jsonFile)

    json.Unmarshal([]byte(byteValue), &root)

    return &root
}

func evaluateInput(remoteClient *RemoteClient, userInput string) *Command {
	if userInput == "" {
		userInput = "welcome"
	}
	fmt.Println(remoteClient.RemoteJID + ":Received userInput cmd: " + userInput)
	commands := root["commands"].(map[string]interface{})
	cmd := Command{}
	var params *UserParams
	var err error
	params , err = readUser(remoteClient)
	userInput = strings.ToLower(userInput)
	for key1, value1 := range commands {
	  if userInput == "vaccine" && params.State != "" && params.District != "" && params.Age > 0 {
	  	// See if we have saved data
	  	if err == nil {
	  		userInput = "loadsaveddata"
	  		fmt.Println(remoteClient.RemoteJID + ":UserInput Changed to loadSavedData")
	  	} else{
	  		fmt.Println(remoteClient.RemoteJID + ":Error while loading data from user file")
	  	}
	  } // Each value is an interface{} type, that is type asserted as a string
	  if key1 == userInput {
	  	value, _ := json.Marshal(value1)
	  	err = json.Unmarshal([]byte(string(value)), &cmd)
	  	if err != nil {
	  		fmt.Println(remoteClient.RemoteJID + ":Error unmarshaling data for " + string(value))
	  	}
	  	break
	  }
	}
	if (remoteClient.LastSent.Name == "welcome" || userInput == "vaccine" || userInput == "loadsaveddata") && cmd.Name == "loadsaveddata" {
		cmd.ToBeSent = fmt.Sprintf(cmd.ToBeSent, params.State, params.District, params.Age, params.BookingPrefs.CenterID, params.BookingPrefs.BookAnySlot)
		fmt.Println(cmd.ToBeSent)
	}
	fmt.Println(remoteClient.RemoteJID + ":Returning cmd:" + cmd.Name)
	return &cmd
}

func readUser(remoteClient *RemoteClient) (*UserParams, error) {
	params := UserParams{}
	path := pkBotFilePath("pkBotUsers_" + remoteClient.RemoteJID + ".json")
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println(remoteClient.RemoteJID + "Error:\n")
		fmt.Println(err)
		writeUser(remoteClient)
		return &params, err
	}
	defer jsonFile.Close()

    byteValue, _ := ioutil.ReadAll(jsonFile)
    err = json.Unmarshal([]byte(byteValue), &params)
    if err != nil {
  		fmt.Println(remoteClient.RemoteJID + ":Error unmarshaling data(readUser) for " + string(byteValue))
  	}
  	bearerToken = params.OTPTxnDetails.BearerToken
    return &params, err
}

func writeUser(remoteClient *RemoteClient) error {
	path := pkBotFilePath("pkBotUsers_" + remoteClient.RemoteJID + ".json")
	jsonFile, err := os.Create(path)
	if err != nil {
		fmt.Println(remoteClient.RemoteJID + ":Error creating file")
		fmt.Println(err)
		return err
	}
	defer jsonFile.Close()
	file, err := json.MarshalIndent(remoteClient.Params, "", " ")
	if err != nil {
		fmt.Println(remoteClient.RemoteJID + ":Error marshaling data for writing")
		fmt.Println(err)
		return err
	}
	err = ioutil.WriteFile(path, file, 0644)
	if err != nil {
		fmt.Println(remoteClient.RemoteJID + ":Error writing data to file")
		fmt.Println(err)
		return err
	}
	return nil
}

func readUsers() (*Subscriptions, error) {
	params := Subscriptions{}
	path := pkBotFilePath("pkBotUsers.json")
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println("Error:\n")
		fmt.Println(err)
		return &params, err
	}
	defer jsonFile.Close()

    byteValue, _ := ioutil.ReadAll(jsonFile)
    err = json.Unmarshal([]byte(byteValue), &params)
    if err != nil {
  		fmt.Println("Error unmarshaling data(readUsers) for " + string(byteValue))
  	}
  	allSubscribers := params.Subscribers
  	allSubscribersMap = make(map[string]*Subscriber)
	for _, subs := range allSubscribers {
		allSubscribersMap[subs.RemoteJID] = &subs
	}
    return &params, err
}

func writeUsers(users *Subscriptions) error {
	path := pkBotFilePath("pkBotUsers.json")
	jsonFile, err := os.Create(path)
	if err != nil {
		fmt.Println("Error creating file:pkBotUsers.json")
		fmt.Println(err)
		return err
	}
	defer jsonFile.Close()
	file, err := json.MarshalIndent(users, "", " ")
	if err != nil {
		fmt.Println("Error marshaling data for writing")
		fmt.Println(err)
		return err
	}
	err = ioutil.WriteFile(path, file, 0644)
	if err != nil {
		fmt.Println("Error writing data to file")
		fmt.Println(err)
		return err
	}
	allSubscribers := users.Subscribers
	allSubscribersMap = make(map[string]*Subscriber)
	for _, subs := range allSubscribers {
		allSubscribersMap[subs.RemoteJID] = &subs
	}
	return nil
}

func pkBotRootDirectoryPath() string {
	path := filepath.Join(os.TempDir(), "pkBot")
	os.MkdirAll(path, os.ModePerm)
	return path
}

func pkBotImageDirectoryPath() string {
	path := filepath.Join(pkBotRootDirectoryPath(), "images")
	os.MkdirAll(path, os.ModePerm)
	return path
}

func pkBotProfilesDirectoryPath() string {
	path := filepath.Join(pkBotRootDirectoryPath(), "profiles")
	os.MkdirAll(path, os.ModePerm)
	return path
}

func pkBotSessionDirectoryPath() string {
	path := filepath.Join(pkBotRootDirectoryPath(), "session")
	os.MkdirAll(path, os.ModePerm)
	return path
}

func pkBotFilePath(fileName string) string {
	path := filepath.Join(pkBotRootDirectoryPath(), fileName)
	fName := strings.ToLower(fileName)
	if strings.HasSuffix(fName, ".jpg") || strings.HasSuffix(fName, ".png") || strings.HasSuffix(fName, ".svg") {
		path = filepath.Join(pkBotImageDirectoryPath(), fileName)
	} else if strings.HasSuffix(fName, ".json") {
		path = filepath.Join(pkBotProfilesDirectoryPath(), fileName)
	} else if strings.HasSuffix(fName, ".gob") {
		path = filepath.Join(pkBotSessionDirectoryPath(), fileName)
	}
	return path
}

func removeSubscriber(s []Subscriber, i int) []Subscriber {
    s[i] = s[len(s)-1]
    // We do not need to put s[i] at the end, as it will be discarded anyway
    return s[:len(s)-1]
}
