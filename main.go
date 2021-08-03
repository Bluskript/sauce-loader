package main

import (
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/Bluskript/sauce-loader/bot"
	"github.com/joho/godotenv"
	"github.com/rapidloop/skv"
)

func chk(e error) {
	if e != nil {
		panic(e)
	}
}

func main() {
	godotenv.Load()

	store, err := skv.Open("./store.db")
	chk(err)

	b, err := bot.New(os.Getenv("BOT_TOKEN"), os.Getenv("TARGET_CHANNEL"), os.Getenv("TARGET_FOLDER"), store)
	chk(err)
	err = b.Open()
	chk(err)

	fmt.Println("Bot is now running.  Press CTRL-C to exit.")
	sc := make(chan os.Signal, 1)
	signal.Notify(sc, syscall.SIGINT, syscall.SIGTERM, os.Interrupt, os.Kill)
	<-sc

	// Cleanly close down the Discord session.
	b.Close()
}
