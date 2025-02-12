package utils

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/ankush-web-eng/contest-backend/models"
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

func SubmitCodeToJudge(code string, languageId int, testCase models.TestCase) (map[string]interface{}, error) {
	url := "https://judge0-ce.p.rapidapi.com/submissions?base64_encoded=false&wait=true&fields=*"
	payload := map[string]interface{}{
		"source_code": code,
		"language_id": languageId,
		"stdin":       testCase.Input,
	}

	jsonData, err := json.Marshal(payload)
	if err != nil {
		return map[string]interface{}{"testCaseId": testCase.ID, "status": "Error"}, err
	}

	req, err := http.NewRequest("POST", url, strings.NewReader(string(jsonData)))
	if err != nil {
		return map[string]interface{}{"testCaseId": testCase.ID, "status": "Error"}, err
	}

	req.Header.Add("x-rapidapi-key", os.Getenv("RAPIDAPI_KEY"))
	req.Header.Add("x-rapidapi-host", "judge0-ce.p.rapidapi.com")
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return map[string]interface{}{"testCaseId": testCase.ID, "status": "Error"}, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return map[string]interface{}{"testCaseId": testCase.ID, "status": "Error"}, err
	}

	return map[string]interface{}{
		"testCaseId":     testCase.ID,
		"problemId":      testCase.ProblemID,
		"status":         result["status"].(map[string]interface{})["description"],
		"stderr":         result["stderr"],
		"stdout":         result["stdout"],
		"timeTaken":      result["time"],
		"memoryUsage":    result["memory"],
		"isHidden":       testCase.IsHidden,
		"expectedOutput": testCase.Output,
	}, nil
}
