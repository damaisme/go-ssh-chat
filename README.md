# SSH Chat Server

This is a simple SSH-based chat server implemented in Go, allowing users to connect via SSH and participate in chat rooms. The server handles user connections, provides basic chat room functionality, and supports commands for interacting with the chat environment.

## Features

- **SSH Login**: Users can connect to the server using SSH without needing a password. Public key authentication is used.
- **Chat Rooms**: Users can join and participate in specific chat rooms. By default, users are placed in the "global" room.
- **Command Support**: Basic commands like `/help` and `/room` allow users to get information and switch chat rooms.
- **Real-time Chat**: Messages in the chat room are broadcast to all users in the same room.

## Prerequisites

1. **Go**: Make sure Go is installed on your machine. You can download it from [golang.org](https://golang.org/).
2. **SSH Private Key**: The server uses an SSH private key to authenticate incoming connections. Ensure you have an RSA private key named `id_rsa` in the same directory as the code or modify the file path as needed.

## Getting Started

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/ssh-chat.git
cd ssh-chat
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Generate SSH host keys
If you don't have an RSA private key already, you can generate one:
```bash
ssh-keygen -t rsa -b 4096 -f id_rs
```

### 4. Build and Run the Server
To build the chat server, run:

```bash
go build -o ssh-chat
./ssh-chat
```
The server will listen on port 2222 by default. You can change this by modifying the net.Listen("tcp", "0.0.0.0:2222") line in the code.

### 5. Connect to the Server
You can connect to the server using any SSH client, for example:

```bash
ssh -p 2222 user@localhost
```
No password is required for login. You will be automatically placed in the global chat room upon connecting.
 
## Commands
- /help: Display a help message with available commands.
- /room <room id>: Switch to a different chat room. If the room does not exist, it will be created.


## Example Usage
```plaintext
user[alice] room[global] => hello everyone!
user[alice] room[global] => /room tech
user[alice] room[tech]   => Let's talk tech!
```

## Code Overview
- Main Entry Point:

- - The main function sets up the SSH server, loads the private key, and listens for incoming connections on port 2222.
SSH Connection Handling:

- The handleConnection function handles the SSH handshake and sets up communication channels for each session.
Chat Room Logic:

- - Each user is associated with a specific chat room. Users can switch rooms using the /room <room id> command.

- Terminal Interaction:

- - The terminal package is used to handle input/output for each user session, simulating a terminal-based chat interface.

## Future Improvements
- Persistent Chat Rooms: Add the ability to persist chat rooms and messages to a database.
- User Authentication: Implement password or key-based authentication for enhanced security.
- Message Broadcast: Improve the broadcast system to ensure messages are delivered in real-time across all connected users.

>> This readme page is generated by gpt
