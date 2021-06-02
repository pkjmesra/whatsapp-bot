// https://pkg.go.dev/github.com/giansalex/whatsapp-go/cl
package pkWhatsApp

import (
	"time"

	"github.com/Rhymen/go-whatsapp"
)

// Message content body whatsapp text message
type Message struct {
	ID string
	From string
	Text string
	Time time.Time
	Source whatsapp.TextMessage
}
