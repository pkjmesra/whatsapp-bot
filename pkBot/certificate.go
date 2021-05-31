package pkBot

import (
	"fmt"
	"os"
)

func downloadcertificate(remoteClient *RemoteClient, cmd *Command) {
	fmt.Println(remoteClient.RemoteJID + ":Now downloading certificate")
	response, err := queryServer(certificateURLFormat, "GET", nil)
	if err != nil {
		fmt.Println(remoteClient.RemoteJID + ":Restarting because failed getting certificate for " + remoteClient.RemoteMobileNumber)
		remoteClient.Host.SendText(remoteClient.RemoteJID, cmd.ErrorResponse2)
		return
	} else {
		certFile, err := os.Create(pkBotFilePath(remoteClient.RemoteMobileNumber + "_Certificate.pdf"))
		if err != nil {
		 	fmt.Println(err)
		}
		defer certFile.Close()
		_ , err = certFile.Write(response)
		if err != nil {
		 	fmt.Println(err)
		}
		fmt.Println(remoteClient.RemoteJID + ":Certificate data saved into :" + pkBotFilePath(remoteClient.RemoteMobileNumber + "_Certificate.pdf"))
	}
}
