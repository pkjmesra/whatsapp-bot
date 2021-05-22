package main

import (
	"github.com/pkjmesra/whatsapp-bot/pkBot"
	"github.com/pkjmesra/whatsapp-bot/pkWhatsApp"
)

func main() {
	remoteClients := make(map[string]*pkBot.RemoteClient)
	client := pkWhatsApp.NewClient()
	pkBot.Initialize()
	client.Listen(func(msg pkWhatsApp.Message) {
		addNewRemoteClient(remoteClients, msg, client)
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
