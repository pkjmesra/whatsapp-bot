package pkBot

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
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

var root map[string]interface{}

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
	fmt.Println("Received userInput cmd: " + userInput)
	commands := root["commands"].(map[string]interface{})
	cmd := Command{}
	var params *UserParams
	var err error
	for key1, value1 := range commands {
	  if strings.ToLower(userInput) == "vaccine" {
	  	// See if we have saved data
	  	params , err = readUser(remoteClient)
	  	if err == nil && remoteClient.LastSent.Name == "welcome" {
	  		userInput = "loadSavedData"
	  		fmt.Println("UserInput Changed to loadSavedData")
	  	}
	  } // Each value is an interface{} type, that is type asserted as a string
	  if key1 == userInput {
	  	value, _ := json.Marshal(value1)
	  	json.Unmarshal([]byte(string(value)), &cmd)
	  	break
	  }
	}
	if remoteClient.LastSent.Name == "welcome" && cmd.Name == "loadSavedData" {
		cmd.ToBeSent = fmt.Sprintf(cmd.ToBeSent, params.State, params.District, params.Age, params.BookingPrefs.CenterID)
		fmt.Println(cmd.ToBeSent)
	}
	return &cmd
}

func readUser(remoteClient *RemoteClient) (*UserParams, error) {
	params := UserParams{}
	path := os.TempDir() + "pkBotUsers_" + remoteClient.RemoteJID + ".json"
	jsonFile, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
		return &params, err
	}
	defer jsonFile.Close()

    byteValue, _ := ioutil.ReadAll(jsonFile)
    json.Unmarshal([]byte(byteValue), &params)

    return &params, err
}

func writeUser(remoteClient *RemoteClient) error {
	path := os.TempDir() + "pkBotUsers_" + remoteClient.RemoteJID + ".json"
	jsonFile, err := os.Create(path)
	if err != nil {
		fmt.Println(err)
		return err
	}
	defer jsonFile.Close()
	file, err := json.MarshalIndent(remoteClient.Params, "", " ")
	if err != nil {
		fmt.Println(err)
		return err
	}
	err = ioutil.WriteFile(path, file, 0644)
	if err != nil {
		return err
	}
	return nil
}
