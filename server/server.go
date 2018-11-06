package main

import (
	"net"
	"os"
	"time"
)

type Song struct {
	audio  os.File
	title  string
	artist string
	tag1   string
	tag2   string
}

type Client struct {
	conn          net.Conn
	username      string
	moderator     bool
	notifications []string
	delay         time.Time
}

type Lobby struct {
	name      string
	admin     Client
	users     []Client
	songQueue []Song
}

func (lobby *Lobby) promoteUser(user Client) {
	for i := 0; i < len(lobby.users); i++ {
		if lobby.users[i].username == user.username {
			lobby.users[i].moderator = true
		}
	}
}

func (lobby *Lobby) addSongToQueue(song Song) {
	lobby.songQueue = append(lobby.songQueue, song)
}

func (lobby *Lobby) getClients() []Client {
	var clients []Client
	clients = append(clients, lobby.admin)
	for _, client := range lobby.users {
		clients = append(clients, client)
	}
	return clients
}

func (client *Client) getDelay() time.Time {

}

func (lobby *Lobby) pushSong(song Song) string {

}

func (lobby *Lobby) syncPlay(song Song) string {

}

func (lobby *Lobby) syncPause() string {

}

func main() {

}
