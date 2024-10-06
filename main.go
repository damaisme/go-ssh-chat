package main

import (
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"strings"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
)

type Chat struct {
	gorm.Model
	Room     string
	Username string
	Time     time.Time
	Message  string `gorm:"size:200"`
}

var chats []Chat

func main() {

	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{})
	if err != nil {
		log.Println("Failed to connect to database:", err)
		return
	}

	err = db.AutoMigrate(&Chat{})
	if err != nil {
		log.Println("Failed to migrate schema:", err)
		return
	}

	result := db.Find(&chats)

	if result.Error != nil {
		log.Println("Error retrieving users:", result.Error)
		return
	}

	fmt.Println(len(chats))

	if len(chats) < 1 {
		chats = append(chats, Chat{Room: "global", Username: "Alice", Time: time.Now(), Message: "Hello World!!!"})
	}

	// SSH server configuration
	config := &ssh.ServerConfig{
		// Add a callback to accept any public key
		PublicKeyCallback: func(c ssh.ConnMetadata, key ssh.PublicKey) (*ssh.Permissions, error) {
			// Accept any public key
			log.Printf("User login: %q", c.User())
			return nil, nil
		},
	}

	// Load the private key
	privateBytes, err := os.ReadFile("id_rsa")
	if err != nil {
		log.Fatalf("Failed to load private key: %v", err)
	}
	private, err := ssh.ParsePrivateKey(privateBytes)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}
	config.AddHostKey(private)

	// Listen on port 2222
	listener, err := net.Listen("tcp", "0.0.0.0:2222")
	if err != nil {
		log.Fatalf("Failed to listen on 2222: %v", err)
	}
	log.Println("Listening on 0.0.0.0:2222...")

	for {
		nConn, err := listener.Accept()
		if err != nil {
			log.Printf("Failed to accept incoming connection: %v", err)
			continue
		}

		go handleConnection(nConn, config, db)
	}
}

func handleConnection(nConn net.Conn, config *ssh.ServerConfig, db *gorm.DB) {
	sshConn, chans, reqs, err := ssh.NewServerConn(nConn, config)

	if err != nil {
		log.Printf("Failed to establish SSH connection: %v", err)
		return
	}
	defer sshConn.Close()

	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			newChannel.Reject(ssh.UnknownChannelType, "unknown channel type")
			continue
		}

		channel, requests, err := newChannel.Accept()
		if err != nil {
			log.Printf("Could not accept channel: %v", err)
			continue
		}

		go handleSession(channel, requests, sshConn.User(), db)
	}
}

func prompt(term *terminal.Terminal, username string, room string, resp string) {
	term.Write([]byte("\033[H\033[2J"))
	// term.Write([]byte("\033[500;0H"))
	if resp == "help" {
		resp = "ssh.dama.lol\n\nAvailable command: \n/help = enter help page \n/room <room id> = select room \n\nAuthor: Damasukma \ngithub: https://github.com/damaisme/ssh-chat"

		term.Write([]byte("Info: " + resp + "\n"))

		prompt := "user[" + username + "] room[" + room + "]" + " => "
		term.SetPrompt(prompt)
		return
	}

	for _, chat := range chats {
		if chat.Room == room {
			message := chat.Time.Format("2006-01-02 15:04:05") + " | " + chat.Username + ": " + chat.Message + "\n"
			term.Write([]byte(message))
		}
	}
	term.Write([]byte("\n\n"))

	if resp != "" {
		term.Write([]byte("Info: " + resp + "\n"))
	}

	prompt := "user[" + username + "] room[" + room + "]" + " => "
	term.SetPrompt(prompt)
	// term.Write([]byte(prompt))

}
func handleSession(channel ssh.Channel, requests <-chan *ssh.Request, username string, db *gorm.DB) {
	defer channel.Close()

	sizeChat := len(chats)

	term := terminal.NewTerminal(channel, "")
	room := "global"
	resp := ""

	prompt(term, username, room, "")
	go func() {
		for {
			if sizeChat != len(chats) {
				time.Sleep(100 * time.Millisecond)

				prompt(term, username, room, resp)
				sizeChat = len(chats)
			}
		}
	}()

	for {
		input, err := term.ReadLine()
		if err != nil {
			if err == io.EOF {
				log.Println("Client disconnected")
				return
			}
			log.Printf("Error reading input: %v", err)
			return
		}

		if input == "" {
			resp = "get help with /help"
			prompt(term, username, room, resp)
		} else if string(input[0]) == "/" {
			parts := strings.Fields(input)

			switch parts[0] {
			case "/help":
				resp = "help"
				prompt(term, username, room, resp)
			case "/room":
				if len(parts) > 1 {
					room = parts[1]
				} else {
					room = "global"
				}
				prompt(term, username, room, resp)
			}
		} else if input != "" {
			resp = ""
			prompt(term, username, room, resp)

			newChat := Chat{Username: username, Room: room, Time: time.Now(), Message: input}
			log.Println(newChat)
			result := db.Create(&newChat)
			chats = append(chats, newChat)

			if result.Error != nil {
				log.Println("Failed to insert new chat:", result.Error)
				return
			}

		} else {
			prompt(term, username, room, resp)
		}

	}
}
