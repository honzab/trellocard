package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
)

type Config struct {
	BoardId  string `json:"board_id"`
	ApiKey   string `json:"api_key"`
	Token    string `json:"token"`
	ListName string `json:"list_name"`
}

type ListsResponse struct {
	Lists []struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	} `json:"lists"`
}

const URL = "https://api.trello.com/1"

var ApiKey string
var Token string
var BoardId string
var ListName string

var configFile = flag.String("config", "trellocard.conf", "Path to the configuration file")

func getListId() (*string, error) {

	v := url.Values{}
	v.Set("lists", "open")
	v.Set("list_fields", "name")
	v.Set("fields", "name,desc")
	v.Set("key", ApiKey)
	v.Set("token", Token)

	getBoardsUrl := fmt.Sprintf("%s/boards/%s?%s", URL, BoardId, v.Encode())

	log.Printf("GET %s", getBoardsUrl)

	resp, err := http.Get(getBoardsUrl)
	if err != nil {
		return nil, err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var lr ListsResponse
	err = json.Unmarshal(body, &lr)
	if err != nil {
		return nil, err
	}

	for i := range lr.Lists {
		if lr.Lists[i].Name == ListName {
			return &lr.Lists[i].Id, nil
		}
	}

	return nil, errors.New("Could not find your list")
}

func createTicket(name, listId string) error {
	cardsUrl := fmt.Sprintf("%s/cards", URL)
	v := url.Values{}
	v.Set("name", name)
	v.Set("pos", "bottom")
	v.Set("idList", listId)
	v.Set("key", ApiKey)
	v.Set("token", Token)

	log.Printf("POST %s %s", cardsUrl, v.Encode())

	response, err := http.PostForm(cardsUrl, v)
	if err != nil {
		return err
	}

	if response.StatusCode == 200 {
		return nil
	} else {
		return errors.New(fmt.Sprintf("got a %d status when creating", response.StatusCode))
	}
}

func loadConfig() error {
	jsonFile, err := os.Open(*configFile)
	if err != nil {
		return err
	}

	config, err := ioutil.ReadAll(jsonFile)
	if err != nil {
		return err
	}
	var conf Config
	err = json.Unmarshal(config, &conf)
	if err != nil {
		return err
	}

	if len(conf.ApiKey) == 0 {
		return errors.New("configuration requires `api_key` to be defined")
	}

	if len(conf.Token) == 0 {
		return errors.New("configuration requires `token` to be defined")
	}

	if len(conf.BoardId) == 0 {
		return errors.New("configuration requires `board_id` to be defined")
	}

	if len(conf.ListName) == 0 {
		return errors.New("configuration requires `list_name` to be defined")
	}

	ApiKey = conf.ApiKey
	Token = conf.Token
	BoardId = conf.BoardId
	ListName = conf.ListName

	return nil
}

func main() {
	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatal("Please specify one parameter: ticket name")
	}

	err := loadConfig()
	if err != nil {
		panic(err)
	}

	listId, err := getListId()
	if err != nil {
		panic(err)
	}

	err = createTicket(flag.Arg(0), *listId)
	if err != nil {
		panic(err)
	}
	fmt.Printf("Created %s\n", flag.Arg(0))
}
