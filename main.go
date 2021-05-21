package main

import (
	"github.com/pkjmesra/whatsapp-bot/whatsapp"
)

func main() {
	client := whatsapp.NewClient()

	client.Listen(func(msg whatsapp.Message) {
		if msg.Text == "Hi" {
			client.SendText(msg.From, "Hello from *github*!")
		}
	})
}
