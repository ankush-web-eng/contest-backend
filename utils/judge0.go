package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func GetLanguageId(language string) (int, error) {
	client := &http.Client{}
	req, err := http.NewRequest("GET", "https://judge029.p.rapidapi.com/languages", nil)
	if err != nil {
		return 0, err
	}

	req.Header.Add("x-rapidapi-key", os.Getenv("RAPIDAPI_KEY"))
	req.Header.Add("x-rapidapi-host", "judge029.p.rapidapi.com")

	resp, err := client.Do(req)
	if err != nil {
		return 0, err
	}
	defer resp.Body.Close()

	var languages []struct {
		ID   int    `json:"id"`
		Name string `json:"name"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&languages); err != nil {
		return 0, err
	}

	for _, lang := range languages {
		if strings.Contains(strings.ToLower(lang.Name), strings.ToLower(language)) {
			return lang.ID, nil
		}
	}

	return 0, fmt.Errorf("language not found")
}
