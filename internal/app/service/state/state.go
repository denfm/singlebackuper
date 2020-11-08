package state

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"time"
)

type State struct {
	UniqueName string
	StatePath  string
}

type Data struct {
	IsBusy        bool
	StatePath     string
	DateTimeLabel string
	DateUnixTime  time.Time
}

type DataFromJSON struct {
	DateUnixTime int64 `json:"DateUnixTime"`
}

func (c *State) GetStatePath() string {
	return fmt.Sprintf("%s%s.state", c.StatePath, c.UniqueName)
}

func (c *State) Clear() {
	statePath := c.GetStatePath()
	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		return
	}

	err := os.Remove(statePath)

	if err != nil {
		log.Fatal(err)
	}
}

func (c *State) GetStateData() Data {
	statePath := c.GetStatePath()
	data := DataFromJSON{}
	timeLocation, err := time.LoadLocation("Europe/Moscow")

	if err != nil {
		log.Fatal(err)
	}

	if _, err := os.Stat(statePath); os.IsNotExist(err) {
		tm := time.Now().In(timeLocation)
		data.DateUnixTime = tm.Unix()
		dataJSON, err := json.Marshal(data)

		if err != nil {
			log.Fatal(err)
		}

		err = ioutil.WriteFile(statePath, dataJSON, 0644)

		if err != nil {
			log.Fatal(err)
		}

		return Data{
			false,
			statePath,
			tm.Format("2006-01-02 15:04:05"),
			tm,
		}
	} else {
		file, err := ioutil.ReadFile(statePath)

		if err != nil {
			log.Fatal(err)
		}

		err = json.Unmarshal(file, &data)

		if err != nil {
			log.Fatal(err)
		}

		tm := time.Unix(data.DateUnixTime, 0)

		return Data{
			true,
			statePath,
			tm.In(timeLocation).Format("2006-01-02 15:04:05"),
			tm,
		}
	}
}

func NewState(uniqueName string, statePath string) *State {
	return &State{uniqueName, statePath}
}
