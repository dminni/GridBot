package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"github.com/dminni/gridbot/models"
)

type BingXClient struct {
	BaseURL string
}

func NewBingXClient(baseURL string) *BingXClient {
	return &BingXClient{
		BaseURL: baseURL,
	}
}

type BingXResponse struct {
	Code int             `json:"code"`
	Msg  string          `json:"msg"`
	Data json.RawMessage `json:"data"`
}

func (c *BingXClient) GetSymbols() (map[string]bool, error) {
	url := fmt.Sprintf("%s/openApi/spot/v1/common/symbols", c.BaseURL)
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var bxResp BingXResponse
	if err := json.NewDecoder(resp.Body).Decode(&bxResp); err != nil {
		return nil, err
	}

	if bxResp.Code != 0 {
		return nil, fmt.Errorf("bingx api error: %s", bxResp.Msg)
	}

	var data struct {
		Symbols []struct {
			Symbol string `json:"symbol"`
		} `json:"symbols"`
	}
	if err := json.Unmarshal(bxResp.Data, &data); err != nil {
		return nil, err
	}

	symbols := make(map[string]bool)
	for _, s := range data.Symbols {
		symbols[s.Symbol] = true
	}

	return symbols, nil
}

func (c *BingXClient) GetOHLCV(symbol string, interval string, limit int) ([]models.OHLCV, error) {
	// BingX uses SYMBOL-USDT format for spot
	bxSymbol := symbol + "-USDT"
	url := fmt.Sprintf("%s/openApi/spot/v1/market/kline?symbol=%s&interval=%s&limit=%d", c.BaseURL, bxSymbol, interval, limit)
	
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var bxResp BingXResponse
	if err := json.NewDecoder(resp.Body).Decode(&bxResp); err != nil {
		return nil, err
	}

	if bxResp.Code != 0 {
		return nil, fmt.Errorf("bingx api error: %s", bxResp.Msg)
	}

	var rawKlines [][]interface{}
	if err := json.Unmarshal(bxResp.Data, &rawKlines); err != nil {
		return nil, err
	}

	var ohlcv []models.OHLCV
	for _, k := range rawKlines {
		if len(k) < 6 {
			continue
		}
		
		ts, _ := k[0].(float64)
		open, _ := strconv.ParseFloat(k[1].(string), 64)
		high, _ := strconv.ParseFloat(k[2].(string), 64)
		low, _ := strconv.ParseFloat(k[3].(string), 64)
		close, _ := strconv.ParseFloat(k[4].(string), 64)
		vol, _ := strconv.ParseFloat(k[5].(string), 64)

		ohlcv = append(ohlcv, models.OHLCV{
			Timestamp: int64(ts),
			Open:      open,
			High:      high,
			Low:       low,
			Close:     close,
			Volume:    vol,
		})
	}

	return ohlcv, nil
}
