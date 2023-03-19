package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
)

type NameInfo struct {
	Name        string 
	CountryID   string  
	Probability float64 
}

var db sync.Map // in-memory


func addNameToDB(name string) {
	resp, err := http.Get(fmt.Sprintf("https://api.nationalize.io?name=%s", name))
	if err != nil {
		fmt.Printf("Error getting nationalize.io data for name %s: %v\n", name, err)
		return
	}
	defer resp.Body.Close()

	var data struct {
		Country []struct {
			CountryID string  `json:"country_id"`
			Prob      float64 `json:"probability"`
		} `json:"country"`
	}

	// json
	err = json.NewDecoder(resp.Body).Decode(&data)
	if err != nil {
		fmt.Printf("Error unmarshalling nationalize.io data for name %s: %v\n", name, err)
		return
	}

	
	var nameInfo []NameInfo
	for _, c := range data.Country {
		nameInfo = append(nameInfo, NameInfo{
			Name:        name,
			CountryID:   c.CountryID,
			Probability: c.Prob,
		})
	}

	
	db.Store(name, nameInfo)
}

func getMinMaxProbabilities(name string) (string, string) {
	
	nameInfo, ok := db.Load(name)
	if !ok {
		addNameToDB(name)
		nameInfo, ok = db.Load(name)
		if !ok {
			return "", ""
		}
	}

	minProb := 1.0
	maxProb := 0.0
	var minCountry, maxCountry string
	for _, info := range nameInfo.([]NameInfo) {
		if info.Probability < minProb {
			minProb = info.Probability
			minCountry = info.CountryID
		}
		if info.Probability > maxProb {
			maxProb = info.Probability
			maxCountry = info.CountryID
		}
	}

	return minCountry, maxCountry
}

func main() {
	names := []string{"Joey", "Aljosja", "Ervinas"}

	for _, name := range names {
		addNameToDB(name)
	}
	for _, name := range names {
		minCountry, maxCountry := getMinMaxProbabilities(name)
		fmt.Printf("Name: %s, Min Probability Country: %s, Max Probability Country: %s\n", name, minCountry, maxCountry)

	}
}
