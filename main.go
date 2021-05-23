package main

import (
	"fmt"
	// "os"
	"strings"

	"github.com/pkjmesra/whatsapp-bot/pkBot"
	"github.com/pkjmesra/whatsapp-bot/pkWhatsApp"
)

func main() {
	remoteClients := make(map[string]*pkBot.RemoteClient)
	client := pkWhatsApp.NewClient()
	pkBot.Initialize()
	client.Listen(func(msg pkWhatsApp.Message) {
		// Only handle messages to self or one-on-one messages to the registered WhatsApp number
		if strings.Contains(msg.From, "@c.us") || strings.Contains(msg.From, "@s.whatsapp.net") {
			addNewRemoteClient(remoteClients, msg, client)
		} else if strings.Contains(msg.From, "@g.us"){
			// Ignore the messages in group
			fmt.Println("Message Received -> ID : " + msg.From + " : Message:" + msg.Text)
		}
	})
}

func addNewRemoteClient(m map[string]*pkBot.RemoteClient, msg pkWhatsApp.Message, wac *pkWhatsApp.WhatsappClient) *pkBot.RemoteClient {
    rc := m[msg.From]
    if rc == nil {
        m[msg.From] = pkBot.NewClient(msg, wac)
        rc = m[msg.From]
    }
    rc.Received = msg
    pkBot.Respond(rc)
    return rc
}
