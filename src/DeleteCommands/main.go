package main

import (
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}
	TOKEN := os.Getenv("TOKEN")

	dg, err := discordgo.New("Bot " + TOKEN)
	if err != nil {
		log.Fatal("Error creating session:", err)
	}

	err = dg.Open()
	if err != nil {
		log.Fatal("Error opening session:", err)
	}

	// グローバルコマンドを削除する
	commands, err := dg.ApplicationCommands(dg.State.User.ID, "")
	if err != nil {
		log.Fatal("Error fetching commands:", err)
	}

	for _, cmd := range commands {
		err := dg.ApplicationCommandDelete(dg.State.User.ID, "", cmd.ID)
		if err != nil {
			log.Println("Failed to delete global command:", cmd.Name, err)
		} else {
			log.Println("Deleted global command:", cmd.Name)
		}
	}

	log.Println("All commands deleted successfully")

	dg.Close()
}
