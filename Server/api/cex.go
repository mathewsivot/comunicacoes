package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"example.com/go/crypto/datatypes"
)

const apiURL = "https://cex.io/api/ticker/%s/USD"

func GetRate(currency string) (*datatypes.Rate, error) {
	upCurrency := strings.ToUpper(currency)
	res, err := http.Get(fmt.Sprintf(apiURL, upCurrency))
	if err != nil {
		return nil, err
	}

	var responseNew CEXResponse
	if res.StatusCode == http.StatusOK {
		bdBytes, err := io.ReadAll(res.Body)
		if err != nil {
			return nil, err
		}

		err = json.Unmarshal(bdBytes, &responseNew)
		if err != nil {
			return nil, err
		}
	} else {
		return nil, fmt.Errorf("StatusCode recebido: %v", res.StatusCode)
	}
	rate := datatypes.Rate{Currency: currency, Price: responseNew.Bid}
	return &rate, nil
}
