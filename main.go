package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
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
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Println("\n=== Name Nationality Probability Checker ===")
		fmt.Println("1. Check a single name")
		fmt.Println("2. Check multiple names")
		fmt.Println("3. Exit")
		fmt.Print("Choose an option (1-3): ")

		choice, _ := reader.ReadString('\n')
		choice = strings.TrimSpace(choice)

		switch choice {
		case "1":
			fmt.Print("Enter a name: ")
			name, _ := reader.ReadString('\n')
			name = strings.TrimSpace(name)

			addNameToDB(name)
			minCountry, maxCountry := getMinMaxProbabilities(name)
			fmt.Printf("\nResults for %s:\n", name)
			fmt.Printf("Most likely country: %s\n", maxCountry)
			fmt.Printf("Least likely country: %s\n", minCountry)

		case "2":
			fmt.Print("Enter names (separated by spaces): ")
			namesInput, _ := reader.ReadString('\n')
			namesInput = strings.TrimSpace(namesInput)
			names := strings.Split(namesInput, " ")

			for _, name := range names {
				addNameToDB(name)
				minCountry, maxCountry := getMinMaxProbabilities(name)
				fmt.Printf("\nResults for %s:\n", name)
				fmt.Printf("Most likely country: %s\n", maxCountry)
				fmt.Printf("Least likely country: %s\n", minCountry)
			}

		case "3":
			fmt.Println("Goodbye!")
			return

		default:
			fmt.Println("Invalid option. Please try again.")
		}
	}
}
