package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"

	"github.com/foolin/pagser"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

func getDistricts() (string, tgbotapi.ReplyKeyboardMarkup) {
	var res string
	for _, district := range regionsData.Districts {
		res += fmt.Sprintf("%v - %v\n", district[0], district[1])
	}

	return res, tgbotapi.NewReplyKeyboard(
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("1"),
			tgbotapi.NewKeyboardButton("2"),
			tgbotapi.NewKeyboardButton("3"),
			tgbotapi.NewKeyboardButton("4"),
		),
		tgbotapi.NewKeyboardButtonRow(
			tgbotapi.NewKeyboardButton("5"),
			tgbotapi.NewKeyboardButton("6"),
			tgbotapi.NewKeyboardButton("7"),
			tgbotapi.NewKeyboardButton("8"),
		),
	)
}

func getDistrictById(id string) (string, error) {

	for _, district := range regionsData.Districts {
		if district[0] == id {
			return district[1], nil
		}
	}
	return "Округ не найден", errors.New("No district")
}

func getRegions(district string) (string, tgbotapi.ReplyKeyboardMarkup) {
	tmp, err := ioutil.ReadFile("data.json")
	if err != nil {
		fmt.Println("src > ", err)
	}

	var data RegionsData
	err = json.Unmarshal(tmp, &data)
	if err != nil {
		fmt.Println("src > ", err)
	}

	d, err := strconv.Atoi(district)
	if err != nil {
		fmt.Println("src > ", err)
	}

	var res string
	var buttons tgbotapi.ReplyKeyboardMarkup
	var row []tgbotapi.KeyboardButton
	for id, region := range data.Regions[d-1] {
		res += fmt.Sprintf("%v - %v\n", region[0], region[1])
		row = append(row, tgbotapi.NewKeyboardButton(region[0]))
		if len(row) >= 5 || len(data.Regions[d-1])-id < 5 {
			buttons.Keyboard = append(buttons.Keyboard, row)
			row = nil
		}
	}

	return res, buttons
}

func getRegionById(id string, district string) (string, error) {
	tmp, err := ioutil.ReadFile("data.json")
	if err != nil {
		fmt.Println("src > ", err)
	}

	var data RegionsData
	err = json.Unmarshal(tmp, &data)
	if err != nil {
		fmt.Println("src > ", err)
	}

	d, err := strconv.Atoi(district)
	if err != nil {
		fmt.Println("src > ", err)
	}

	regions := data.Regions[d-1]

	for _, region := range regions {
		if region[0] == id {
			return region[1], nil
		}
	}
	return "", errors.New("Регион не найден")
}

func getAlertRegionsData(region string) PageData {

	var resp *http.Response
	var err error
	var url string

	if len(region) < 3 {
		url = "https://meteoinfo.ru/informer/meteoalert/?a=0" + region
	} else {
		url = "https://meteoinfo.ru/informer/meteoalert/?a=" + region
	}

	client := &http.Client{}
	client.Timeout = time.Minute * 3

	req, err := http.NewRequest("GET", url, nil)

	if err != nil {
		fmt.Println("src > ", err)
	}

	resp, err = client.Do(req)
	if err != nil {
		fmt.Println("src > ", err)
	}

	defer resp.Body.Close()
	html, err := io.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("src > ", err)
	}

	p := pagser.New()
	var data PageData

	err = p.Parse(&data, string(html))
	if err != nil {
		fmt.Println("src > ", err)
	}

	return data
}

func delayedAlert(data AlertTimer) {
	for true {
		t := time.Now()
		if t.Hour() == 13 && t.Minute() == 10 {

			msg := tgbotapi.NewMessage(data.ChatID, "")

			content := getAlertRegionsData(data.Region)
			fmt.Println(content)

			var events = "Оповещения:"
			for _, event := range content.Events {
				events += "\n" + event
			}
			msg.Text = content.Region[1] + "\n\n" + events

			if _, err := Bot.Send(msg); err != nil {
				fmt.Println("src > ", err)
			}
		}
		time.Sleep(55 * time.Second)
	}

}
