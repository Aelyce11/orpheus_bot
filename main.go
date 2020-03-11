package main

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/bwmarrin/dgvoice"
	"github.com/bwmarrin/discordgo"
)

var (
	botID    string
	prefix   = ";"
	voiceCon *discordgo.VoiceConnection
)

func main() {

	// Start bot
	bot := generateBot("TOKEN")

	// Add handlers
	fmt.Println("Registering handlers .. ")
	bot.AddHandler(ready)
	bot.AddHandler(messageCreate)

	fmt.Println("Opening connection ..")
	err := bot.Open()
	check("Error opening connection to Discord", err)

	// Wait until interrupt
	fmt.Println("Orpheus is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleaning up
	fmt.Println("Cleaning up ..")
	err = bot.Close()
	check("Bot couldn't close properly", err)
}

func generateBot(token string) *discordgo.Session {
	fmt.Println("Generating bot ..")
	bot, err := discordgo.New("Bot " + token)
	check("Error creating Dirscord session", err)
	fmt.Println("Bot generated ..")
	return bot
}

func ready(session *discordgo.Session, event *discordgo.Ready) {
	fmt.Println("Handling 'ready' state ..")
	err := session.UpdateStatus(0, "Orpheus")
	check("Status couldn't be set", err)
	servers := session.State.Guilds
	fmt.Printf("Orpheus has started on %d servers\n", len(servers))
	fmt.Println(event.SessionID)
}

func messageCreate(session *discordgo.Session, message *discordgo.MessageCreate) {
	user := message.Author
	if user.ID == session.State.User.ID || user.Bot {
		//Do nothing because the bot is talking
		return
	}

	content := message.Content

	if strings.HasPrefix(content, prefix) {
		content = strings.TrimPrefix(content, prefix)
		if strings.HasPrefix(content, "join") {
			voiceCon = joinVoiceChannel(session, message)
		}
		if strings.HasPrefix(content, "leave") {
			leaveVoiceChannel(voiceCon)
		}
		if strings.HasPrefix(content, "mp3") {
			filename := strings.TrimPrefix(content, "mp3 ")
			playMp3(filename, session, message.ChannelID)
		}
	}
}

func joinVoiceChannel(session *discordgo.Session, message *discordgo.MessageCreate) *discordgo.VoiceConnection {
	channel, err := session.State.Channel(message.ChannelID)
	check("Couldn't get message source channel", err)
	guild, err := session.State.Guild(channel.GuildID)
	check("Couldn't find guild", err)

	// Join voice channel
	for _, voiceState := range guild.VoiceStates {
		if voiceState.UserID == message.Author.ID {
			voiceChannel, err := session.ChannelVoiceJoin(guild.ID, voiceState.ChannelID, false, false)
			check("Couldn't join voice channel", err)
			return voiceChannel
		}
	}
	return nil
}

func leaveVoiceChannel(voiceCon *discordgo.VoiceConnection) {
	err := voiceCon.Disconnect()
	check("couldn't disconnect from voice channel", err)
}

func playMp3(msg string, session *discordgo.Session, channel string) {

	fmt.Println("Reading Folder: ", "C:\\Users\\alice\\Music\\"+msg)
	files, err := ioutil.ReadDir("C:\\Users\\alice\\Music\\" + msg)
	check("couldn't read folder", err)
	for _, f := range files {
		fmt.Println("PlayAudioFile:", f.Name())
		session.ChannelMessageSend(channel, "Playing : "+strings.TrimSuffix(f.Name(), ".mp3")+" by "+msg)

		dgvoice.PlayAudioFile(voiceCon, fmt.Sprintf("%s/%s", "C:\\Users\\alice\\Music\\"+msg, f.Name()), make(chan bool))
	}

	return
}

func check(msg string, err error) {
	if err != nil {
		fmt.Println(msg, err)
	}
}
