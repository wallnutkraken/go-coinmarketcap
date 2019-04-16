// Package coinmarketcap Coin Market Cap API client for Go
package coinmarketcap

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"
	"time"
)

// Interface interface
type Interface interface {
	CryptocurrencyInfo(options *CryptocurrencyInfoOptions) (map[string]*CryptocurrencyInfo, error)
	CryptocurrencyListingsLatest(options *CryptocurrencyListingsLatestOptions) ([]*Listing, error)
}

// Status is the status structure
type Status struct {
	Timestamp    string  `json:"timestamp"`
	ErrorCode    int     `json:"error_code"`
	ErrorMessage *string `json:"error_message"`
	Elapsed      int     `json:"elapsed"`
	CreditCount  int     `json:"credit_count"`
}

// Response is the response structure
type Response struct {
	Status Status      `json:"status"`
	Data   interface{} `json:"data"`
}

// Listing is the listing structure
type Listing struct {
	ID                float64           `json:"id"`
	Name              string            `json:"name"`
	Symbol            string            `json:"symbol"`
	Slug              string            `json:"slug"`
	CirculatingSupply float64           `json:"circulating_supply"`
	TotalSupply       float64           `json:"total_supply"`
	MaxSupply         float64           `json:"max_supply"`
	DateAdded         string            `json:"date_added"`
	NumMarketPairs    float64           `json:"num_market_pairs"`
	CMCRank           float64           `json:"cmc_rank"`
	LastUpdated       string            `json:"last_updated"`
	Quote             map[string]*Quote `json:"quote"`
}

// Quote is the quote structure
type Quote struct {
	Price            float64 `json:"price"`
	Volume24H        float64 `json:"volume_24h"`
	PercentChange1H  float64 `json:"percent_change_1h"`
	PercentChange24H float64 `json:"percent_change_24h"`
	PercentChange7D  float64 `json:"percent_change_7d"`
	MarketCap        float64 `json:"market_cap"`
	LastUpdated      string  `json:"last_updated"`
}

// CryptocurrencyInfo options
type CryptocurrencyInfo struct {
	ID       float64                `json:"id"`
	Name     string                 `json:"name"`
	Symbol   string                 `json:"symbol"`
	Category string                 `json:"category"`
	Slug     string                 `json:"slug"`
	Logo     string                 `json:"logo"`
	Tags     []string               `json:"tags"`
	Urls     map[string]interface{} `json:"urls"`
}

// CryptocurrencyInfoOptions options
type CryptocurrencyInfoOptions struct {
	ID     string
	Symbol string
}

// CryptocurrencyListingsLatestOptions options
type CryptocurrencyListingsLatestOptions struct {
	Start   int
	Limit   int
	Convert string
	Sort    string
}

// SortOptions sort options
var SortOptions sortOptions

type sortOptions struct {
	Name              string
	Symbol            string
	DateAdded         string
	MarketCap         string
	Price             string
	CirculatingSupply string
	TotalSupply       string
	MaxSupply         string
	NumMarketPairs    string
	Volume24H         string
	PercentChange1H   string
	PercentChange24H  string
	PercentChange7D   string
}

type MarketQuoteOptions struct {
	ID      string
	Symbol  string
	Convert string
}

type MarketQuoteInfo struct {
	ID                int              `json:"id"`
	Name              string           `json:"name"`
	Symbol            string           `json:"symbol"`
	Slug              string           `json:"slug"`
	CirculatingSupply float64              `json:"circulating_supply"`
	TotalSupply       float64              `json:"total_supply"`
	MaxSupply         float64              `json:"max_supply"`
	DateAdded         time.Time        `json:"date_added"`
	NumMarketPairs    int              `json:"num_market_pairs"`
	CmcRank           int              `json:"cmc_rank"`
	LastUpdated       time.Time        `json:"last_updated"`
	Tags              []string         `json:"tags"`
	Quote             map[string]Quote `json:"quote"`
}

// Client the CoinMarketCap client
type Client struct {
	proAPIKey string
}

// Config the client config structure
type Config struct {
	ProAPIKey string
}

var (
	// ErrCouldNotCast could not cast error
	ErrCouldNotCast = errors.New("could not cast")
)

var (
	siteURL               = "https://coinmarketcap.com"
	baseURL               = "https://pro-api.coinmarketcap.com/v1"
	coinGraphURL          = "https://graphs2.coinmarketcap.com/currencies"
	globalMarketGraphURL  = "https://graphs2.coinmarketcap.com/global/marketcap-total"
	altcoinMarketGraphURL = "https://graphs2.coinmarketcap.com/global/marketcap-altcoin"
)

// NewClient initializes a new client
func NewClient(cfg *Config) *Client {
	if cfg == nil {
		cfg = new(Config)
	}

	if cfg.ProAPIKey == "" {
		cfg.ProAPIKey = os.Getenv("CMC_PRO_API_KEY")
	}

	if cfg.ProAPIKey == "" {
		log.Fatal("Pro API Key is required")
	}

	return &Client{
		proAPIKey: cfg.ProAPIKey,
	}
}

// MarketQuote gets the latest market quote for 1 or more cryptocurrencies.
// Use the "convert" option to return market values in multiple fiat and cryptocurrency conversions in the same call.
func (s *Client) MarketQuote(options *MarketQuoteOptions) (map[string]MarketQuoteInfo, error) {
	var params []string
	if options == nil {
		options = &MarketQuoteOptions{}
	}
	if options.ID != "" {
		params = append(params, fmt.Sprintf("id=%s", options.ID))
	}
	if options.Symbol != "" {
		params = append(params, fmt.Sprintf("symbol=%s", options.Symbol))
	}
	if options.Convert != "" {
		params = append(params, fmt.Sprintf("convert=%s", options.Convert))
	}

	url := fmt.Sprintf("%s/cryptocurrency/quotes/latest?%s", baseURL, strings.Join(params, "&"))
	body, err := s.makeReq(url)
	resp := new(Response)
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	var result = make(map[string]MarketQuoteInfo)
	ifcs, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, ErrCouldNotCast
	}

	for k, v := range ifcs {
		info := MarketQuoteInfo{}
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(b, &info)
		if err != nil {
			return nil, err
		}
		result[k] = info
	}

	return result, nil
}

// CryptocurrencyInfo returns all static metadata for one or more cryptocurrencies including name, symbol, logo, and its various registered URLs.
func (s *Client) CryptocurrencyInfo(options *CryptocurrencyInfoOptions) (map[string]*CryptocurrencyInfo, error) {
	var params []string
	if options == nil {
		options = new(CryptocurrencyInfoOptions)
	}
	if options.ID != "" {
		params = append(params, fmt.Sprintf("id=%s", options.ID))
	}
	if options.Symbol != "" {
		params = append(params, fmt.Sprintf("symbol=%s", options.Symbol))
	}

	url := fmt.Sprintf("%s/cryptocurrency/info?%s", baseURL, strings.Join(params, "&"))

	body, err := s.makeReq(url)
	resp := new(Response)
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, err
	}

	var result = make(map[string]*CryptocurrencyInfo)
	ifcs, ok := resp.Data.(map[string]interface{})
	if !ok {
		return nil, ErrCouldNotCast
	}

	for k, v := range ifcs {
		info := new(CryptocurrencyInfo)
		b, err := json.Marshal(v)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(b, info)
		if err != nil {
			return nil, err
		}
		result[k] = info
	}

	return result, nil
}

// CryptocurrencyListingsLatest gets a paginated list of all cryptocurrencies with latest market data. You can configure this call to sort by market cap or another market ranking field. Use the "convert" option to return market values in multiple fiat and cryptocurrency conversions in the same call.
func (s *Client) CryptocurrencyListingsLatest(options *CryptocurrencyListingsLatestOptions) ([]*Listing, error) {
	var params []string
	if options == nil {
		options = new(CryptocurrencyListingsLatestOptions)
	}
	if options.Start != 0 {
		params = append(params, fmt.Sprintf("start=%v", options.Start))
	}
	if options.Limit != 0 {
		params = append(params, fmt.Sprintf("limit=%v", options.Limit))
	}
	if options.Convert != "" {
		params = append(params, fmt.Sprintf("convert=%s", options.Convert))
	}
	if options.Sort != "" {
		params = append(params, fmt.Sprintf("sort=%s", options.Sort))
	}

	url := fmt.Sprintf("%s/cryptocurrency/listings/latest?%s", baseURL, strings.Join(params, "&"))

	body, err := s.makeReq(url)
	resp := new(Response)
	err = json.Unmarshal(body, &resp)
	if err != nil {
		return nil, fmt.Errorf("JSON Error: [%s]. Response body: [%s]", err.Error(), string(body))
	}

	var listings []*Listing
	ifcs, ok := resp.Data.([]interface{})
	if !ok {
		return nil, ErrCouldNotCast
	}

	for i := range ifcs {
		ifc := ifcs[i]
		listing := new(Listing)
		b, err := json.Marshal(ifc)
		if err != nil {
			return nil, err
		}
		err = json.Unmarshal(b, listing)
		if err != nil {
			return nil, err
		}
		listings = append(listings, listing)
	}

	return listings, nil
}

func init() {
	SortOptions = sortOptions{
		Name:              "name",
		Symbol:            "symbol",
		DateAdded:         "date_added",
		MarketCap:         "market_cap",
		Price:             "price",
		CirculatingSupply: "circulating_supply",
		TotalSupply:       "total_supply",
		MaxSupply:         "max_supply",
		NumMarketPairs:    "num_market_pairs",
		Volume24H:         "volume_24h",
		PercentChange1H:   "percent_change_1h",
		PercentChange24H:  "percent_change_24h",
		PercentChange7D:   "percent_change_7d",
	}
}
