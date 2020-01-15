package feed

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"

	"github.com/kava-labs/kava/kvoracle/types"
)

const (
	coinGeckoBaseURL = "https://api.coingecko.com/api/v3/coins/"
)

// GeckoPrices gets prices for an array of coins by their symbols
func GeckoPrices(symbols []string, convert string) []types.Asset {
	var assets []types.Asset
	client := &http.Client{}

	for _, symbol := range symbols {
		// Format URL and HTTP request
		coinURL := fmt.Sprintf("%s/%s/tickers", coinGeckoBaseURL, symbol)
		req, err := http.NewRequest("GET", coinURL, nil)
		if err != nil {
			log.Print(err)
			os.Exit(1)
		}
		q := url.Values{}
		q.Add("convert", convert)
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
		var data *types.CoinGeckoTickers
		err = json.Unmarshal(respBody, &data)
		if err != nil {
			fmt.Println(err)
		}

		// Use coin's USDT market from Binance
		if data != nil && data.Tickers != nil {
			for _, ticker := range data.Tickers {
				if ticker.Market.Name == "Binance" && ticker.Target == "USDT" {
					asset := types.Asset{
						Symbol: data.Name,
						Price:  ticker.Last,
					}
					assets = append(assets, asset)
				}
			}
		}
	}
	return assets
}
