package user

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
)

func CheckPaymentStatus(userid string, httpClient *http.Client) (bool, error) {
	checkURL := os.Getenv("PAY_API_URL")
	if checkURL == "" {
		checkURL = "https://pay.shortentrack.com"
	}

	url := fmt.Sprintf("%s/check/%s", checkURL, userid)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return false, err
	}

	passcode := os.Getenv("CHECK_PASSCODE")
	if passcode == "" {
		return false, errors.New("missing passcode")
	}
	req.Header.Set("X-Passcode-ID", passcode)

	resp, err := httpClient.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == 400 || resp.StatusCode == 500 {
		return false, errors.New("server error")
	}

	if !strings.Contains(resp.Header.Get("Content-Type"), "application/json") {
		return false, errors.New("unexpected content type, expected JSON")
	}

	var result struct {
		ID     string `json:"id"`
		Paying bool   `json:"paying"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return false, err
	}

	return result.Paying, nil
}
