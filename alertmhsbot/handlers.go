package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"strconv"

	"github.com/bradfitz/gomemcache/memcache"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

// userStatuses:
// 		start
// 		set district
// 		set region
// 		get updates

func receivedMessageHandler(chatId string, userStatus string, receivedMessage string) {
	//fmt.Println("handler > rm ", receivedMessage, userStatus)
	chatId64, err := strconv.ParseInt(chatId, 10, 64)
	if err != nil {
		panic(err)
	}
	msg := tgbotapi.NewMessage(chatId64, "")
	if receivedMessage == "🆘 помощь" {
		msg.Text = "при ЧП звоните на номер 101, 112"
	}
	if receivedMessage == "Свердловская обл." {
		msg.Text =
			"В свердловская обл. часто встречаются: " +
				"\n " +
				"1. Шквал - резкое кратковременное усиление ветра в течение не менее 1 мин. Максимальная скорость ветра (порыв) 25 м/с и более." +
				"\n " +
				"2. Сильный ливень" +
				"\n " +
				"3. Крупный град."
	}
	if receivedMessage == "🌏 сменить регион" {
		err := mc.Set(&memcache.Item{Key: strconv.Itoa(int(chatId64)), Value: NewSessionData("set district", "0", "0")})
		if err != nil {
			fmt.Println("handler > ", err)
		}

		text, buttons := getDistricts()
		msg.Text, msg.ReplyMarkup = "Выберите свой округ\n"+text, buttons
	}

	switch userStatus {
	case "start":
		err := mc.Set(&memcache.Item{Key: strconv.Itoa(int(chatId64)), Value: NewSessionData("set district", "0", "0")})
		if err != nil {
			fmt.Println("handler > ", err)
		}

		text, buttons := getDistricts()
		msg.Text, msg.ReplyMarkup = "Бот создан для оповещений о ЧС 🚨 и опастных 🌦️ погодных  условиях\nВыберите округ\n"+text, buttons
		fmt.Println("handler > ", "start")
	case "set district":
		fmt.Println("handler > ", "set district")
		i, err := strconv.Atoi(receivedMessage)
		if err != nil || i > 8 || i < 1 {
			fmt.Println("handler > ", err)
			msg.Text = "Введите номер округа"
			if _, err = Bot.Send(msg); err != nil {
				fmt.Println("handler > ", err)
			}
			return
		}
		err = mc.Set(&memcache.Item{Key: strconv.Itoa(int(chatId64)), Value: NewSessionData("set region", receivedMessage, "0")})
		if err != nil {
			fmt.Println("handler > ", err)
			msg.Text = "Упс, кажется что-то пошло не так!"
			if _, err = Bot.Send(msg); err != nil {
				fmt.Println("handler > ", err)
			}
			return
		}
		//fmt.Println("handler > rm ", receivedMessage)

		text, buttons := getRegions(receivedMessage)
		msg.Text, msg.ReplyMarkup = "Выберите ваш регион\n"+text, buttons
	case "set region":
		fmt.Println("handler > ", "set region")
		i, err := strconv.Atoi(receivedMessage)
		if err != nil || i < 1 || i > 111 {
			fmt.Println("handler > ", err)
			msg.Text = "Введите номер региона"
			if _, err = Bot.Send(msg); err != nil {
				fmt.Println("handler > ", err)
			}
			return
		}

		it, err := mc.Get(strconv.Itoa(int(chatId64)))
		var data SessionData
		if err = json.Unmarshal(it.Value, &data); err != nil {
			fmt.Println("handler > ", err)
			msg.Text = "Упс, кажется что-то пошло не так!"
			if _, err = Bot.Send(msg); err != nil {
				fmt.Println("handler > ", err)
			}
			return
		}

		err = mc.Set(&memcache.Item{Key: it.Key, Value: NewSessionData("get updates", data.DistrictId, receivedMessage)})
		if err != nil {
			fmt.Println("handler > ", err)
			msg.Text = "Упс, кажется что-то пошло не так!"
			if _, err = Bot.Send(msg); err != nil {
				fmt.Println("handler > ", err)
			}
			return
		}

		text, err := getRegionById(receivedMessage, data.DistrictId)
		if err != nil {
			msg.Text = err.Error()
		} else {
			msg.Text = "Вы будете получать ежедневные оповещения о погодных условиях для " + text +
				"\n " +
				"\nЧтобы получить оповещение сейчас нажмите '🔔 получить уведомление'" +
				"\n " +
				"\nЧтобы сбросить выбраный регион используйте '🌏 сменить регион'"
				//"\nЧтобы узнать интересный факт о погоде используйте /funfact"
			msg.ReplyMarkup = tgbotapi.NewReplyKeyboard(
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("🔔 получить уведомление"),
					tgbotapi.NewKeyboardButton("🌏 сменить регион"),
					//tgbotapi.NewKeyboardButton("/funfact"),
				),
				tgbotapi.NewKeyboardButtonRow(
					tgbotapi.NewKeyboardButton("🆘 помощь"),
					tgbotapi.NewKeyboardButton(text),
					//tgbotapi.NewKeyboardButton("/funfact"),
				),
			)
		}

		at := AlertTimer{
			chatId64,
			receivedMessage,
		}
		go delayedAlert(at)
	case "get updates":
		fmt.Println("handler > ", "get updates")
		if receivedMessage == "🔔 получить уведомление" {
			it, err := mc.Get(strconv.Itoa(int(chatId64)))
			if err != nil {
				fmt.Println("handler > ", err)
				msg.Text = "Упс, кажется что-то пошло не так!"
				if _, err = Bot.Send(msg); err != nil {
					fmt.Println("handler > ", err)
				}
				return
			}

			var data SessionData
			if err = json.Unmarshal(it.Value, &data); err != nil {
				fmt.Println("handler > ", err)
				msg.Text = "Упс, кажется что-то пошло не так!"
				if _, err = Bot.Send(msg); err != nil {
					fmt.Println("handler > ", err)
				}
				return
			}

			content := getAlertRegionsData(data.RegionId)

			var events = "Оповещения:"
			for _, event := range content.Events {
				events += "\n" + event
			}
			msg.Text = content.Region[1] + "\n" + events
		}
		if receivedMessage == "/funfact" {
			id := rand.Intn(5)
			tmp, err := ioutil.ReadFile("funfacts.json")
			if err != nil {
				fmt.Println("handler > ", err)
				msg.Text = "Упс, кажется что-то пошло не так!"
				if _, err = Bot.Send(msg); err != nil {
					fmt.Println("handler > ", err)
				}
				return
			}
			text := ParseFunFacts(tmp)
			msg.Text = text.Facts[id]
			if _, err = Bot.Send(msg); err != nil {
				fmt.Println("handler > ", err)
			}
			photo := tgbotapi.NewPhoto(chatId64, tgbotapi.FilePath("images/"+strconv.Itoa(id)+".jpg"))
			if _, err = Bot.Send(photo); err != nil {
				log.Fatalln(err)
			}
			return
		}
	}

	if m, err := Bot.Send(msg); err != nil {
		fmt.Println(m, err)
	}
}
