// https://pkg.go.dev/github.com/giansalex/whatsapp-go/cl
package pkWhatsApp

import (
	"encoding/gob"
	"fmt"
    "os"
    "path/filepath"
    "strings"
	// "os/signal"
	"time"

	qrcodeTerminal "github.com/Baozisoftware/qrcode-terminal-go"
	"github.com/Rhymen/go-whatsapp"
)

var (loggedInWANumber string)

// WhatsappClient connect
type WhatsappClient struct {
	wac 			*whatsapp.Conn
	MobileNumber 	string
}

// NewClient create whatsapp client
func NewClient() *WhatsappClient {
	wac, err := whatsapp.NewConn(60 * time.Second)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error creating connection: %v\n", err)
		os.Exit(1)
		return nil
	}

	wac.SetClientName("CoWIN Whatsapp Bot", "CoWIN Bot", "0.4.2080")
	wac.SetClientVersion(0, 4, 2080)
	err = login(wac)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error logging in: %v\n", err)
		os.Exit(1)
		return nil
	}

	fmt.Println("Whatsapp Connected!")
	return &WhatsappClient{wac:wac, MobileNumber:loggedInWANumber}
}

// Listen text messages
func (wp *WhatsappClient) Listen(f messageListener) {
	wp.wac.AddHandler(&messageHandler{f, time.Now().Unix()})

	// sigs := make(chan os.Signal, 1)
	// signal.Notify(sigs, os.Interrupt)

	// fmt.Println("Press ctrl+c to exit.")

	// <-sigs
	// fmt.Println("Shutdown.")
	// os.Exit(0)
}

// SendText send text message
func (wp *WhatsappClient) SendText(to string, text string) {
	reply := whatsapp.TextMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: to,
		},
		Text: text,
	}

	msgID, err := wp.wac.Send(reply)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending message: %v", err)
	} else {
		fmt.Println("Message Sent -> ID : " + msgID)
	}
}

func (wp *WhatsappClient) SendImage(to, imageFile, extension, caption string) {
	f, err := os.Open(imageFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error while loading %s image:%v",extension, err)
	}
	defer f.Close()
	reply := whatsapp.ImageMessage{
		Info: whatsapp.MessageInfo{
			RemoteJid: to,
		},
		Type:    "image/" + extension, // png needs to be changed if you used different extension
		Caption: caption,
		Content: f,
	}

	msgID, err := wp.wac.Send(reply)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error sending message: %v", err)
	} else {
		fmt.Println("Message Sent -> ID : " + msgID)
	}
}

// msg := whatsapp.VideoMessage{
// 	Info: whatsapp.MessageInfo{
// 		RemoteJid: "###########@s.whatsapp.net",
// 	},
// 	Type:    "video/mp4",
// 	Caption: "Hello Gopher!",
// 	Content: video, // the code os.Open("video.mp4")
// }

// msg := whatsapp.AudioMessage{
// 	Info: whatsapp.MessageInfo{
// 		RemoteJid: "###########@s.whatsapp.net",
// 	},
// 	Type:    "audio/ogg; codecs=opus",
// 	Content: audio, // the code os.Open("audio.ogg")
// }

// msg := whatsapp.AudioMessage{
// 	Info: whatsapp.MessageInfo{
// 		RemoteJid: "###########@s.whatsapp.net",
// 	},
// 	Type:    "audio/mp3; codecs=opus", // tried removing `; codecs=opus` but nothing happens
// 	Content: audio, // os.Open("audio.mp3")
// }

// msg := whatsapp.DocumentMessage{
// 	Info: whatsapp.MessageInfo{
// 		RemoteJid: "###########@s.whatsapp.net",
// 	},
// 	Type:      "document/docx",
// 	Thumbnail: thumbnail,
// 	Content:   document, // os.Open("document.docx")
// }

// GetConnection return whatsapp connection
func (wp *WhatsappClient) GetConnection() *whatsapp.Conn {
	return wp.wac
}

func login(wac *whatsapp.Conn) error {
	//load saved session
	session, err := readSession()
	if err == nil {
		//restore session
		session, err = wac.RestoreWithSession(session)
		if err != nil {
			return fmt.Errorf("restoring failed: %v", err)
		}
	} else {
		//no saved session -> regular login
		qr := make(chan string)
		go func() {
			terminal := qrcodeTerminal.New()
			terminal.Get(<-qr).Print()
		}()
		session, err = wac.Login(qr)
		if err != nil {
			return fmt.Errorf("error during login: %v", err)
		}
		fmt.Println("Login done -> ID : " + string(session.Wid))
	}

	//save session
	err = writeSession(session)
	if err != nil {
		return fmt.Errorf("error saving session: %v", err)
	}
	loggedInWANumber = strings.TrimSuffix(string(session.Wid), "@c.us")
	return nil
}

func readSession() (whatsapp.Session, error) {
	session := whatsapp.Session{}
	file, err := os.Open(pkBotFilePath("whatsappSession.gob"))
	if err != nil {
		return session, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&session)
	if err != nil {
		return session, err
	}
	return session, nil
}

func writeSession(session whatsapp.Session) error {
	file, err := os.Create(pkBotFilePath("whatsappSession.gob"))
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	err = encoder.Encode(session)
	if err != nil {
		return err
	}
	return nil
}

func pkBotRootDirectoryPath() string {
	path := filepath.Join(os.TempDir(), "pkBot")
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
	if strings.HasSuffix(fName, ".gob") {
		path = filepath.Join(pkBotSessionDirectoryPath(), fileName)
	}
	return path
}
