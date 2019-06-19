package main

import (
	"encoding/json"
	"fmt"
	"gopkg.in/telegram-bot-api.v4"
	"io/ioutil"
	"log"
	"net/http"
	"os"
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
	fmt.Println(configuration.TelegramBotToken)
	return configuration.TelegramBotToken
}

func initBot () {
	bot, err := tgbotapi.NewBotAPI(loadBotConfig())
	if err != nil {
		log.Panic(err)
	}
	bot.Debug = true
	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60
	updates, err := bot.GetUpdatesChan(u)
	oldCollection := Apartment{}
	oldCollection = getApartments(&oldCollection)
	newCollection := Apartment{}
	oldMap := make(apartmentsIds)
	newMap := make(apartmentsIds)
	for update := range updates {
		switch {
			case update.Message.Text == "/get":

				message := getNewApartments(&newCollection, &oldMap, newMap)
				fmt.Println("ya tut")
				fmt.Println(len(message))
				for id := range message {
					if len(message) != 0 {
						msg := newCollection.Apartments[id].URL
						test := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
						_, err := bot.Send(test)
						if err != nil {
							panic(err)
						}
					} else {
						msg := "There is no new apartments!"
						test := tgbotapi.NewMessage(update.Message.Chat.ID, msg)
						_, err := bot.Send(test)
						if err != nil {
							panic(err)
						}
					}
				}
				fmt.Println(oldMap)
				fmt.Println(newMap)
				oldMap = newMap
				fmt.Println(oldMap)
				fmt.Println(newMap)
				oldCollection = newCollection
		    case update.Message.Text == "/help":
		    	s := "/get - get list of new aparments in Minsk city"
				msg := tgbotapi.NewMessage(update.Message.Chat.ID, s)
				_, err := bot.Send(msg)
				if err != nil {
					panic(err)
				}
		}
	}
}

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

func getApartments(c *Apartment) Apartment {
	url := "https://ak.api.onliner.by/search/apartments?rent_type%5B%5D=1_room&only_owner=true&price%5Bmin%5D=290&price%5Bmax%5D=600&currency=usd&bounds%5Blb%5D%5Blat%5D=53.69914561462634&bounds%5Blb%5D%5Blong%5D=27.36625671386719&bounds%5Brt%5D%5Blat%5D=54.09604689032579&bounds%5Brt%5D%5Blong%5D=27.75833129882813&page=1&v=0.6142670626784548"
	spaceClient := http.Client{
		Timeout: time.Second * 2, // Maximum of 2 secs
	}
	req, err := http.NewRequest(http.MethodGet, url, nil)
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
	return *c
}

func getNewApartments (nc *Apartment, om *apartmentsIds, nm apartmentsIds) []int {
	*nc = getApartments(nc)
	var diffApartments []int
	initApartments(*nc, nm)
	for id := range nm {
		if _, lol := (*om)[id] ; !lol {
			diffApartments = append(diffApartments, id)
		}
	}
	return diffApartments
}

func initApartments (c Apartment, m apartmentsIds) {
	for i := range c.Apartments {
		m[c.Apartments[i].ID] = struct{}{}
	}
}

type apartmentsIds map[int]struct{}


func main() {
	initBot()
}