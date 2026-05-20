package ibapi

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// ContractService handles contract and security definition API calls.
type ContractService struct {
	client *Client
}

// SecDef represents a security definition.
type SecDef struct {
	Conid          int     `json:"conid"`
	Currency       string  `json:"currency"`
	Time           int     `json:"time"`
	ChineseName    string  `json:"chineseName"`
	AllExchanges   string  `json:"allExchanges"`
	ListingExchange string `json:"listingExchange"`
	CountryCode    string  `json:"countryCode"`
	Name           string  `json:"name"`
	AssetClass     string  `json:"assetClass"`
	Expiry         string  `json:"expiry,omitempty"`
	LastTradingDay string  `json:"lastTradingDay,omitempty"`
	Group          string  `json:"group"`
	PutOrCall      string  `json:"putOrCall,omitempty"`
	Sector         string  `json:"sector"`
	SectorGroup    string  `json:"sectorGroup"`
	Strike         string  `json:"strike"`
	Ticker         string  `json:"ticker"`
	UndConid       int     `json:"undConid"`
	Multiplier     float64 `json:"multiplier"`
	Type           string  `json:"type"`
	HasOptions     bool    `json:"hasOptions"`
	FullName       string  `json:"fullName"`
	IsUS           bool    `json:"isUS"`
	IncrementRules []struct {
		LowerEdge float64 `json:"lowerEdge"`
		Increment float64 `json:"increment"`
	} `json:"incrementRules"`
	DisplayRule struct {
		Magnification int `json:"magnification"`
		DisplayRuleStep []struct {
			DecimalDigits int     `json:"decimalDigits"`
			LowerEdge     float64 `json:"lowerEdge"`
			WholeDigits   int     `json:"wholeDigits"`
		} `json:"displayRuleStep"`
	} `json:"displayRule"`
	IsEventContract bool `json:"isEventContract"`
	PageSize        int  `json:"pageSize"`
}

// SecDefSearchResult represents the response from /iserver/secdef/search.
type SecDefSearchResult struct {
	Conid          string   `json:"conid"`
	CompanyHeader  string   `json:"companyHeader"`
	CompanyName    string   `json:"companyName"`
	Symbol         string   `json:"symbol"`
	Description    string   `json:"description,omitempty"`
	Restricted     string   `json:"restricted,omitempty"`
	Sections       []struct {
		SecType  string `json:"secType"`
		Months   string `json:"months,omitempty"`
		Exchange string `json:"exchange,omitempty"`
	} `json:"sections"`
	Fop    interface{} `json:"fop,omitempty"`
	Opt    interface{} `json:"opt,omitempty"`
	War    interface{} `json:"war,omitempty"`
	Issuers []struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		BondID  int    `json:"bondid"`
		Conid   string `json:"conid"`
		CompanyHeader string `json:"companyHeader"`
		CompanyName   interface{} `json:"companyName"`
		Symbol        interface{} `json:"symbol"`
		Description   interface{} `json:"description"`
		Restricted    interface{} `json:"restricted"`
	} `json:"issuers,omitempty"`
}

// SecDefInfo represents contract info returned by /iserver/secdef/info.
type SecDefInfo struct {
	Conid          int     `json:"conid"`
	Symbol         string  `json:"symbol"`
	SecType        string  `json:"secType"`
	Exchange       string  `json:"exchange"`
	ListingExchange interface{} `json:"listingExchange"`
	Right         string  `json:"right,omitempty"`
	Strike         float64 `json:"strike"`
	Currency       string  `json:"currency"`
	Cusip          interface{} `json:"cusip"`
	Coupon         string  `json:"coupon"`
	Desc1          string  `json:"desc1"`
	Desc2          string  `json:"desc2"`
	MaturityDate   string  `json:"maturityDate"`
	Multiplier     string  `json:"multiplier"`
	TradingClass   string  `json:"tradingClass"`
	ValidExchanges string  `json:"validExchanges"`
}

// ContractInfo represents detailed contract information from /iserver/contract/{conid}/info.
type ContractInfo struct {
	CFI_CODE            string  `json:"cfi_code"`
	Symbol              string  `json:"symbol"`
	Cusip               interface{} `json:"cusip"`
	ExpiryFull          interface{} `json:"expiry_full"`
	ConID               int     `json:"con_id"`
	MaturityDate        string  `json:"maturity_date"`
	Industry            string  `json:"industry"`
	InstrumentType      string  `json:"instrument_type"`
	TradingClass        string  `json:"trading_class"`
	ValidExchanges      string  `json:"valid_exchanges"`
	AllowSellLong       bool    `json:"allow_sell_long"`
	IsZeroCommissionSecurity bool `json:"is_zero_commission_security"`
	LocalSymbol         string  `json:"local_symbol"`
	ContractClarificationType interface{} `json:"contract_clarification_type"`
	Classifier          interface{} `json:"classifier"`
	Currency            string  `json:"currency"`
	Text                interface{} `json:"text"`
	UnderlyingConID     int     `json:"underlying_con_id"`
	RTH                 bool    `json:"r_t_h"`
	Multiplier          interface{} `json:"multiplier"`
	UnderlyingIssuer    interface{} `json:"underlying_issuer"`
	ContractMonth       interface{} `json:"contract_month"`
	CompanyName         string  `json:"company_name"`
	SmartAvailable      bool    `json:"smart_available"`
	Exchange            string  `json:"exchange"`
	Category            string  `json:"category"`
}

// ConidInfo represents a contract identifier with exchange.
type ConidInfo struct {
	Conid   int    `json:"conid"`
	Exchange string `json:"exchange"`
}

// Ticker represents ticker information from /trsrv/stocks.
type Ticker struct {
	Conid    int    `json:"conid"`
	Exchange string `json:"exchange"`
	IsUS     bool   `json:"isUS"`
}

// StockContract represents stock contract information.
type StockContract struct {
	Name         string      `json:"name"`
	ChineseName  interface{} `json:"chineseName"`
	AssetClass   string      `json:"assetClass"`
	Contracts    []Ticker   `json:"contracts"`
}

// FuturesContract represents a futures contract.
type FuturesContract struct {
	Symbol           string `json:"symbol"`
	Conid            int    `json:"conid"`
	UnderlyingConid  int    `json:"underlyingConid"`
	ExpirationDate   int    `json:"expirationDate"`
	LTD             int    `json:"ltd"`
	ShortFuturesCutOff int `json:"shortFuturesCutOff"`
	LongFuturesCutOff  int  `json:"longFuturesCutOff"`
}

// TradingSchedule represents a trading schedule entry.
type TradingSchedule struct {
	ClearingCycleEndTime string `json:"clearingCycleEndTime"`
	TradingScheduleDate string `json:"tradingScheduleDate"`
	Sessions            []struct {
		OpeningTime string `json:"openingTime"`
		ClosingTime string `json:"closingTime"`
		Prop        string `json:"prop,omitempty"`
	} `json:"sessions"`
	TradingTimes []struct {
		Description string `json:"description,omitempty"`
		OpeningTime string `json:"openingTime"`
		ClosingTime string `json:"closingTime"`
		CancelDayOrders string `json:"cancelDayOrders,omitempty"`
	} `json:"tradingtimes,omitempty"`
}

// ScheduleResponse represents the response from /trsrv/secdef/schedule.
type ScheduleResponse struct {
	ID            string             `json:"id"`
	TradeVenueID  string             `json:"tradeVenueId"`
	Timezone      string             `json:"timezone,omitempty"`
	Schedules     []TradingSchedule  `json:"schedules"`
}

// TradingScheduleDetail represents detailed trading schedule.
type TradingScheduleDetail struct {
	ExchangeTimeZone string `json:"exchange_time_zone"`
	Schedules       map[string]struct {
		ExtendedHours []struct {
			CancelDailyOrders bool `json:"cancel_daily_orders"`
			Closing           int  `json:"closing"`
			Opening           int  `json:"opening"`
		} `json:"extended_hours"`
		LiquidHours []struct {
			Closing int `json:"closing"`
			Opening int `json:"opening"`
		} `json:"liquid_hours"`
	} `json:"schedules"`
}

// BondFilters represents bond filter options.
type BondFilters struct {
	DisplayText string `json:"displayText"`
	ColumnID    int    `json:"columnId"`
	Options     []struct {
		Text  string `json:"text,omitempty"`
		Value string `json:"value"`
	} `json:"options"`
}

// SecdefSearch searches for contracts by symbol.
// Set name=true to search by company name instead of ticker symbol.
func (s *ContractService) SecdefSearch(ctx context.Context, symbol string, opts ...SecdefSearchOption) ([]SecDefSearchResult, error) {
	params := url.Values{}
	params.Set("symbol", symbol)

	for _, opt := range opts {
		opt(params)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.BuildURL("/v1/api/iserver/secdef/search?"+params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var results []SecDefSearchResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return nil, err
	}

	return results, nil
}

// SecdefSearchOption is a functional option for secdef search.
type SecdefSearchOption func(url.Values)

// WithNameSearch enables company name search mode.
func WithNameSearch() SecdefSearchOption {
	return func(params url.Values) {
		params.Set("name", "true")
	}
}

// WithSecType filters by security type (STK, IND, BOND).
func WithSecType(secType string) SecdefSearchOption {
	return func(params url.Values) {
		params.Set("secType", secType)
	}
}

// SecdefInfo returns contract details for a specific conid.
// Required params: conid, secType, month
// Optional: exchange, strike (for options), right (for options), issuerId (for bonds)
func (s *ContractService) SecdefInfo(ctx context.Context, conid int, secType, month string, opts ...SecdefInfoOption) (*SecDefInfo, error) {
	params := url.Values{}
	params.Set("conid", strconv.Itoa(conid))
	params.Set("secType", secType)
	params.Set("month", month)

	for _, opt := range opts {
		opt(params)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.BuildURL("/v1/api/iserver/secdef/info?"+params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result SecDefInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SecdefInfoOption is a functional option for secdef info.
type SecdefInfoOption func(url.Values)

// WithExchange specifies the exchange for the contract.
func WithExchange(exchange string) SecdefInfoOption {
	return func(params url.Values) {
		params.Set("exchange", exchange)
	}
}

// WithStrike specifies the strike price for options/futures options.
func WithStrike(strike string) SecdefInfoOption {
	return func(params url.Values) {
		params.Set("strike", strike)
	}
}

// WithRight specifies the right (C for Call, P for Put) for options.
func WithRight(right string) SecdefInfoOption {
	return func(params url.Values) {
		params.Set("right", right)
	}
}

// WithIssuerId specifies the issuer ID for bonds.
func WithIssuerId(issuerId string) SecdefInfoOption {
	return func(params url.Values) {
		params.Set("issuerId", issuerId)
	}
}

// ContractInfo returns detailed contract information by conid.
func (s *ContractService) ContractInfo(ctx context.Context, conid int) (*ContractInfo, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.BuildURL(fmt.Sprintf("/v1/api/iserver/contract/%d/info", conid)), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ContractInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SecdefStrikes returns available strike prices for an underlying contract.
func (s *ContractService) SecdefStrikes(ctx context.Context, conid int, secType, month string, opts ...SecdefStrikesOption) (*StrikesResponse, error) {
	params := url.Values{}
	params.Set("conid", strconv.Itoa(conid))
	params.Set("secType", secType)
	params.Set("month", month)

	for _, opt := range opts {
		opt(params)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.BuildURL("/v1/api/iserver/secdef/strikes?"+params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result StrikesResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// SecdefStrikesOption is a functional option for secdef strikes.
type SecdefStrikesOption func(url.Values)

// WithStrikesExchange specifies the exchange for the contract.
func WithStrikesExchange(exchange string) SecdefStrikesOption {
	return func(params url.Values) {
		params.Set("exchange", exchange)
	}
}

// StrikesResponse represents the response from /iserver/secdef/strikes.
type StrikesResponse struct {
	Call []float64 `json:"call"`
	Put  []float64 `json:"put"`
}

// SecdefByConids returns security definitions for given conids.
// Accepts a comma-separated list of conids.
func (s *ContractService) SecdefByConids(ctx context.Context, conids []int) ([]SecDef, error) {
	ids := make([]string, len(conids))
	for i, id := range conids {
		ids[i] = strconv.Itoa(id)
	}
	params := url.Values{}
	params.Set("conids", strings.Join(ids, ","))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.BuildURL("/v1/api/trsrv/secdef?"+params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		Secdef []SecDef `json:"secdef"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Secdef, nil
}

// AllConidsByExchange returns all contracts available on a given exchange.
func (s *ContractService) AllConidsByExchange(ctx context.Context, exchange string) ([]ConidInfo, error) {
	params := url.Values{}
	params.Set("exchange", exchange)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.BuildURL("/v1/api/trsrv/all-conids?"+params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []ConidInfo
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// Stocks returns stock contracts for given symbols.
// Symbols must be comma-separated and contain only capitalized letters.
func (s *ContractService) Stocks(ctx context.Context, symbols []string) (map[string][]StockContract, error) {
	params := url.Values{}
	params.Set("symbols", strings.Join(symbols, ","))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.BuildURL("/v1/api/trsrv/stocks?"+params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string][]StockContract
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// Futures returns futures contracts for given symbols.
func (s *ContractService) Futures(ctx context.Context, symbols []string) (map[string][]FuturesContract, error) {
	params := url.Values{}
	params.Set("symbols", strings.Join(symbols, ","))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.BuildURL("/v1/api/trsrv/futures?"+params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string][]FuturesContract
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// TradingSchedule retrieves the trading schedule for a contract.
func (s *ContractService) TradingSchedule(ctx context.Context, assetClass, conid, symbol string, opts ...TradingScheduleOption) (*ScheduleResponse, error) {
	params := url.Values{}
	params.Set("assetClass", assetClass)
	params.Set("conid", conid)
	params.Set("symbol", symbol)

	for _, opt := range opts {
		opt(params)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.BuildURL("/v1/api/trsrv/secdef/schedule?"+params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ScheduleResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// TradingScheduleOption is a functional option for trading schedule.
type TradingScheduleOption func(url.Values)

// WithScheduleExchange specifies the primary exchange.
func WithScheduleExchange(exchange string) TradingScheduleOption {
	return func(params url.Values) {
		params.Set("exchange", exchange)
	}
}

// WithExchangeFilter specifies the exchange to retrieve data from.
func WithExchangeFilter(exchange string) TradingScheduleOption {
	return func(params url.Values) {
		params.Set("exchangeFilter", exchange)
	}
}

// TradingScheduleNew retrieves the trading schedule for the 6 days surrounding the current trading day.
func (s *ContractService) TradingScheduleNew(ctx context.Context, conid, exchange string) (*TradingScheduleDetail, error) {
	params := url.Values{}
	params.Set("conid", conid)
	if exchange != "" {
		params.Set("exchange", exchange)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.BuildURL("/v1/api/contract/trading-schedule?"+params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result TradingScheduleDetail
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// BondFilters retrieves bond filter information.
func (s *ContractService) BondFilters(ctx context.Context, symbol, issuerId string) ([]BondFilters, error) {
	params := url.Values{}
	params.Set("symbol", symbol)
	params.Set("issuerId", issuerId)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.BuildURL("/v1/api/iserver/secdef/bond-filters?"+params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result struct {
		BondFilters []BondFilters `json:"bondFilters"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.BondFilters, nil
}