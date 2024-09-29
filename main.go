package main

import (
	"io"
	"log"
	"net"
	"os"
	"time"

	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/terminal"
	"strings"
)

type Chat struct {
	room     string
	username string
	time     time.Time
	value    string
}

var chats []Chat

func main() {

	chats = append(chats, Chat{username: "alice", room: "global", time: time.Now(), value: "hallooo"})
	chats = append(chats, Chat{username: "bob", room: "global", time: time.Now(), value: "haiii"})

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

		go handleConnection(nConn, config)
	}
}

func handleConnection(nConn net.Conn, config *ssh.ServerConfig) {
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

		go handleSession(channel, requests, sshConn.User())
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
		if chat.room == room {
			message := chat.time.Format("2006-01-02 15:04:05") + " | " + chat.username + ": " + chat.value + "\n"
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
func handleSession(channel ssh.Channel, requests <-chan *ssh.Request, username string) {
	defer channel.Close()

	term := terminal.NewTerminal(channel, "")
	room := "global"
	resp := ""

	prompt(term, username, room, "")
	go func() {
		for {
			time.Sleep(200 * time.Millisecond)

			prompt(term, username, room, resp)
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
				log.Println("len :" + string(len(parts)))
				if len(parts) > 0 {
					room = parts[1]
				} else {
					room = "global"
				}
				prompt(term, username, room, resp)
			}
		} else if input != "" {
			resp = ""
			prompt(term, username, room, resp)
			chats = append(chats, Chat{username: username, room: room, time: time.Now(), value: input})
		} else {
			prompt(term, username, room, resp)
		}

	}
}
