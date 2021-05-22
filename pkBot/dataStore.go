package pkBot

import (
    "encoding/json"
    "fmt"
    "io/ioutil"
    "os"
)

type Command struct {
	Name      			string  `json:"name"`
	CommandType         string  `json:"commandType"`
	ToBeSent     		string  `json:"toBeSent"`
	ResponseType    	string  `json:"responseType"`
	ExpectedResponse    string  `json:"expectedResponse"`
	NextCommand  		string  `json:"nextCommand"`
	ErrorResponse		string 	`json:"errorResponse"`
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

func evaluateInput(userInput string) *Command {
	if userInput == "" {
		userInput = "welcome"
	}
	commands := root["commands"].(map[string]interface{})
	cmd := Command{}
	for key1, value1 := range commands {
	  // Each value is an interface{} type, that is type asserted as a string
	  if key1 == userInput {
	  	value, _ := json.Marshal(value1)
	  	json.Unmarshal([]byte(string(value)), &cmd)
	  	break
	  }
	}
	return &cmd
}
