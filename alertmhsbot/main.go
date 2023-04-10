package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/joho/godotenv"
)

var (
	Bot         *tgbotapi.BotAPI
	regionsData RegionsData
	//updates     tgbotapi.UpdatesChannel
	mc *memcache.Client
)

func init() {
	fmt.Println("Reading env file")
	err := godotenv.Load(".env")
	if err != nil {
		panic(err)
	}

	token := os.Getenv("TOKEN")
	if len(token) > 0 {
		fmt.Println("Token received")
	} else {
		fmt.Println("Token didnt received")
	}

	Bot, err = tgbotapi.NewBotAPI(token)
	if err != nil {
		log.Fatalln(err)
	}
	//Bot.Debug = true
	log.Printf("Authorized on account %s", Bot.Self.UserName)

	tmp, err := ioutil.ReadFile("data.json")
	if err != nil {
		fmt.Println(err)
	}

	err = json.Unmarshal(tmp, &regionsData)
	if err != nil {
		fmt.Println(err)
	}

	mc = memcache.New("127.0.0.1:11211") // "127.0.0.1:11211", "localhost:11211"
	if err = mc.Ping(); err != nil {
		panic(err)
	}
	err = mc.DeleteAll()
	if err != nil {
		fmt.Println(err)
	}
}

func main() {
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates := Bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {
			fmt.Println("main > ", update.Message.Text)
			if it, err := mc.Get(strconv.Itoa(int(update.Message.Chat.ID))); it == nil {
				err = mc.Set(&memcache.Item{Key: strconv.Itoa(int(update.Message.Chat.ID)), Value: NewSessionData("start", "0", "0")})
				if err != nil {
					panic(err)
				}
				receivedMessageHandler(strconv.Itoa(int(update.Message.Chat.ID)), "start", update.Message.Text)
			} else if it != nil {
				p := ParseSessionData(it.Value)
				fmt.Println("main > ", it.Key, string(it.Value))
				receivedMessageHandler(it.Key, p.Status, update.Message.Text)
			} else {
				fmt.Println("main > err >", err)
			}

		}
	}

}
