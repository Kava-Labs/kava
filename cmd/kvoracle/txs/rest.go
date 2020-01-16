package txs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/kava-labs/kava/cmd/kvoracle/types"
)

const (
	BaseURL = "http://localhost:1317"
)

// GetAssetList gets a list of assets on kava
func GetAssetList() {
	client := &http.Client{}

	// Format URL and HTTP request
	requestURL := fmt.Sprintf("%s/%s", BaseURL, "pricefeed/assets")
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	q := url.Values{}
	req.Header.Set("Accepts", "application/json")
	req.URL.RawQuery = q.Encode()

	// Make an HTTP request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request to server")
		os.Exit(1)
	}

	// Read the response
	respBody, _ := ioutil.ReadAll(resp.Body)

	// Unmarshal the response to a usable format
	var data *types.MarketsRes
	err = json.Unmarshal(respBody, &data)
	if err != nil {
		fmt.Println(err)
	}

	// TODO: Unmarshal to object instead of string
	for _, asset := range data.Result {
		fmt.Println(asset)
	}
}

// GetAssetPrice gets an asset's current price on kava
func GetAssetPrice(symbol string) {
	client := &http.Client{}

	// Format URL and HTTP request
	requestURL := fmt.Sprintf("%s/%s/%s", BaseURL, "pricefeed/currentprice/", symbol)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		log.Print(err)
		os.Exit(1)
	}
	q := url.Values{}
	req.Header.Set("Accepts", "application/json")
	req.URL.RawQuery = q.Encode()

	// Make an HTTP request
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Error sending request to server")
		os.Exit(1)
	}

	// Read the response
	respBody, _ := ioutil.ReadAll(resp.Body)

	fmt.Println(string(respBody))
}
