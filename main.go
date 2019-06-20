package main

import (
	"bytes"
	"encoding/json"
	"github.com/robfig/cron"
	"gopkg.in/telegram-bot-api.v4"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"
)

type Config struct {
	TelegramBotToken string
}

func loadBotConfig() string {
	file, _ := os.Open("config.json")
	decoder := json.NewDecoder(file)
	configuration := Config{}
	err := decoder.Decode(&configuration)
	if err != nil {
		log.Panic(err)
	}
	return configuration.TelegramBotToken
}

func addChannel(c *channels, channelId int64) {
	(*c)[channelId] = struct{}{}
}

func DeleteChannel(c *channels, channelId int64) {
	delete(*c, channelId)
}

func initBot () {
	channels := make(channels)
	bot, err := tgbotapi.NewBotAPI(loadBotConfig())
	if err != nil {
		log.Panic(err)
	}
	//bot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	var url = "https://ak.api.onliner.by/search/apartments?rent_type%5B%5D=1_room&rent_type%5B%5D=2_rooms&price%5Bmin%5D=300&price%5Bmax%5D=600&currency=usd&only_owner=true&bounds%5Blb%5D%5Blat%5D=53.69914561462634&bounds%5Blb%5D%5Blong%5D=27.36625671386719&bounds%5Brt%5D%5Blat%5D=54.09604689032579&bounds%5Brt%5D%5Blong%5D=27.75833129882813&v=0.18898162215768832"
	c := Apartment{}
	oldMap := make(apartmentsIds)

	cr := cron.New()
	_ = cr.AddFunc("*/1 * * * * *", func() {
		message, diff := getNewApartments(&url, &c, &oldMap)
		if len(diff) != 0 {
			for channel := range channels {
				test := tgbotapi.NewMessage(channel, message)
				_, err := bot.Send(test)
				if err != nil {
					panic(err)
				}
			}
		}
	})
	cr.Start()

	for update := range updates {
		switch {
			case update.Message.Text == "/get":
				message, diff := getNewApartments(&url, &c, &oldMap)
				if len(diff) != 0 {
					for channel := range channels {
						test := tgbotapi.NewMessage(channel, message)
						_, err := bot.Send(test)
						if err != nil {
							panic(err)
						}
					}
				} else {
						msg := "There is no new apartments!"
						test := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
						_, err := bot.Send(test)
						if err != nil {
							panic(err)
						}
					}
		    case update.Message.Text == "/help":
		    	s := "/get - get list of new aparments in Minsk city \n/channelid - get your channel id \n/start - register \n/exit - unregister"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, s)
				_, err := bot.Send(msg)
				if err != nil {
					panic(err)
				}
			case update.Message.Text == "/channelid":
				s := strconv.FormatInt(update.Message.Chat.ID, 10)
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, s)
				_, err := bot.Send(msg)
				if err != nil {
					panic(err)
				}
			case update.Message.Text == "/start":
				addChannel(&channels, update.Message.Chat.ID)
				s := "You are registered for updates!"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, s)
				_, err := bot.Send(msg)
				if err != nil {
					panic(err)
				}
			case update.Message.Text == "/exit":
				DeleteChannel(&channels, update.Message.Chat.ID)
				s := "You are unregistered from updates!"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, s)
				_, err := bot.Send(msg)
				if err != nil {
					panic(err)
				}
		}
	}
}

type apartmentsIds map[int]struct{}

type channels map[int64]struct{}

type Apartment struct {
	Apartments []struct {
		ID    int `json:"id"`
		Price struct {
			Amount    string `json:"amount"`
			Converted struct {
				USD struct {
					Amount   string `json:"amount"`
				} `json:"USD"`
			} `json:"converted"`
		} `json:"price"`
		RentType string `json:"rent_type"`
		Location struct {
			Address     string  `json:"address"`
		} `json:"location"`
		Photo   string `json:"photo"`
		Contact struct {
			Owner bool `json:"owner"`
		} `json:"contact"`
		CreatedAt     string `json:"created_at"`
		LastTimeUp    string `json:"last_time_up"`
		URL           string `json:"url"`
	} `json:"apartments"`
}

func getApartments(c *Apartment, url *string) apartmentsIds {
	m := make(apartmentsIds)
	spaceClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	req, err := http.NewRequest(http.MethodGet, *url, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.Header.Set("Accept", "application/json")
	resp, getErr := spaceClient.Do(req)
	if getErr != nil {
		log.Fatal(getErr)
	}
	bodyBytes, readErr := ioutil.ReadAll(resp.Body)
	if readErr != nil {
		log.Fatal(readErr)
	}
	err = json.Unmarshal(bodyBytes, &c)
	if err != nil {
		panic(err)
	}
	for i := range c.Apartments {
		m[c.Apartments[i].ID] = struct{}{}
	}
	return m
}

func getNewApartments (url *string, c *Apartment, om *apartmentsIds) (string, []int) {
	nm := getApartments(c, url)
	var diffApartments []int
	var buffer bytes.Buffer
	for key, _ := range nm {
		_, ok := (*om)[key]
		if !ok {
			diffApartments = append(diffApartments, key)
		}
	}
	for _, id := range diffApartments {
		for _, a := range c.Apartments {
			if a.ID == id {
				buffer.WriteString(a.URL + "\n")
			}
		}
	}
	*om = nm
	return buffer.String(), diffApartments
}

func main() {
	initBot()
}