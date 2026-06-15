package ibapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"
)

// MarketDataService handles market data API calls.
type MarketDataService struct {
	client *Client
}

// HistoricalDataRequest represents request parameters for historical data.
type HistoricalDataRequest struct {
	Conid     string
	Exchange  string
	Period    string
	Bar       string
	StartTime string
	OutsideRth bool
	Source    string
}

// HistoricalDataBar represents a single bar of historical data.
type HistoricalDataBar struct {
	O float64 `json:"o"`
	C float64 `json:"c"`
	H float64 `json:"h"`
	L float64 `json:"l"`
	V float64 `json:"v"`
	T int64   `json:"t"`
}

// HistoricalDataResponse represents the response from /iserver/marketdata/history.
type HistoricalDataResponse struct {
	ServerID        string             `json:"serverId"`
	Symbol          string             `json:"symbol"`
	Text            string             `json:"text"`
	PriceFactor     string             `json:"priceFactor,string"`
	StartTime       string             `json:"startTime"`
	High            string             `json:"high"`
	Low             string             `json:"low"`
	TimePeriod      string             `json:"timePeriod"`
	BarLength       int                `json:"barLength"`
	MDAvailability  string             `json:"mdAvailability"`
	MktDataDelay    int                `json:"mktDataDelay"`
	OutsideRth      bool               `json:"outsideRth"`
	TradingDayDuration int             `json:"tradingDayDuration,omitempty"`
	VolumeFactor    int                `json:"volumeFactor"`
	PriceDisplayRule int               `json:"priceDisplayRule"`
	PriceDisplayValue string           `json:"priceDisplayValue"`
	NegativeCapable  bool               `json:"negativeCapable"`
	MessageVersion  int                `json:"messageVersion"`
	Data            []HistoricalDataBar `json:"data"`
	Points          int                `json:"points"`
	TravelTime      int                `json:"travelTime"`
	Direction       int                `json:"direction,omitempty"`
	ChartPanStartTime string          `json:"chartPanStartTime,omitempty"`
}

// MarketDataSnapshotRequest represents request parameters for market data snapshot.
type MarketDataSnapshotRequest struct {
	Conids []int
	Fields []string
}

// MarketDataSnapshot represents market data for a single contract.
type MarketDataSnapshot struct {
	Updated   int64             `json:"_updated"`
	ConidEx   string            `json:"conidEx"`
	Conid     int              `json:"conid"`
	ServerID  string            `json:"server_id"`
	Fields    map[string]string `json:"-"`
}

// UnmarshalJSON implements custom JSON unmarshaling for MarketDataSnapshot
// to handle dynamic field keys.
func (s *MarketDataSnapshot) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	s.Fields = make(map[string]string)
	for k, v := range raw {
		switch k {
		case "_updated":
			if v != nil {
				s.Updated = int64(v.(float64))
			}
		case "conidEx":
			if v != nil {
				s.ConidEx = v.(string)
			}
		case "conid":
			if v != nil {
				s.Conid = int(v.(float64))
			}
		case "server_id", "6119":
			if v != nil {
				s.ServerID = v.(string)
			}
		default:
			if v != nil {
				s.Fields[k] = toString(v)
			}
		}
	}
	return nil
}

// toString converts various types to string representation.
func toString(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case float64:
		return strconv.FormatFloat(val, 'f', -1, 64)
	case bool:
		return strconv.FormatBool(val)
	default:
		return ""
	}
}

// MarketDataField IDs for snapshot requests.
// Full list available in the IB Web API documentation.
var MarketDataFields = map[string]string{
	"31":  "LastPrice",
	"55":  "Symbol",
	"58":  "Text",
	"70":  "High",
	"71":  "Low",
	"73":  "MarketValue",
	"74":  "AvgPrice",
	"75":  "UnrealizedPnL",
	"76":  "FormattedPosition",
	"77":  "FormattedUnrealizedPnL",
	"78":  "DailyPnL",
	"79":  "RealizedPnL",
	"80":  "UnrealizedPnLPercent",
	"82":  "Change",
	"83":  "ChangePercent",
	"84":  "BidPrice",
	"85":  "AskSize",
	"86":  "AskPrice",
	"87":  "Volume",
	"88":  "BidSize",
	"201": "Right",
}

// HistoricalData retrieves historical market data for a contract.
func (s *MarketDataService) HistoricalData(ctx context.Context, req HistoricalDataRequest) (*HistoricalDataResponse, error) {
	params := url.Values{}
	params.Set("conid", req.Conid)
	if req.Exchange != "" {
		params.Set("exchange", req.Exchange)
	}
	params.Set("period", req.Period)
	params.Set("bar", req.Bar)
	if req.StartTime != "" {
		params.Set("startTime", req.StartTime)
	}
	if req.OutsideRth {
		params.Set("outsideRth", "true")
	}
	if req.Source != "" {
		params.Set("source", req.Source)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.BuildURL("/v1/api/iserver/marketdata/history?"+params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.doRequest(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result HistoricalDataResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Snapshot retrieves market data snapshot for given conids.
func (s *MarketDataService) Snapshot(ctx context.Context, conids []int, fields []string) ([]MarketDataSnapshot, error) {
	params := url.Values{}

	ids := make([]string, len(conids))
	for i, id := range conids {
		ids[i] = strconv.Itoa(id)
	}
	params.Set("conids", joinInts(ids))
	params.Set("fields", joinStrings(fields))

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.BuildURL("/v1/api/iserver/marketdata/snapshot?"+params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.doRequest(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result []MarketDataSnapshot
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// MarketDataAvailability codes.
const (
	MDARealTime         = "R" // Real-time data, market data subscription required
	MDADelayed          = "D" // Delayed 15-20 min
	MDAFrozen           = "Z" // Frozen (last recorded at market close), real-time
	MDAFrozenDelayed    = "Y" // Frozen, delayed
	MDANotSubscribed    = "N" // Not subscribed
	MDAIncompleteAPIAck = "i" // Incomplete Market Data API Acknowledgement
)

// SnapshotField represents a market data field that can be requested.
type SnapshotField struct {
	ID          string
	Name        string
	Description string
}

// Common snapshot fields for stocks.
var CommonStockFields = []string{
	"31",  // Last Price
	"84",  // Bid Price
	"86",  // Ask Price
	"85",  // Ask Size
	"88",  // Bid Size
	"70",  // High
	"71",  // Low
	"87",  // Volume
	"82",  // Change
	"83",  // Change %
	"55",  // Symbol
	"58",  // Text (company name)
}

// MarketDataSnapshotForConid is a helper that creates a single-conid snapshot request.
func (s *MarketDataService) SnapshotForConid(ctx context.Context, conid int, fields []string) (*MarketDataSnapshot, error) {
	snapshots, err := s.Snapshot(ctx, []int{conid}, fields)
	if err != nil {
		return nil, err
	}
	if len(snapshots) == 0 {
		return nil, nil
	}
	return &snapshots[0], nil
}

// Unsubscribe sends an unsubscribe request for market data.
func (s *MarketDataService) Unsubscribe(ctx context.Context, conid int) error {
	body := map[string]interface{}{
		"conid": conid,
	}

	return s.unsubscribe(ctx, body)
}

// UnsubscribeAll cancels all market data requests.
func (s *MarketDataService) UnsubscribeAll(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.BuildURL("/v1/api/iserver/marketdata/unsubscribeall"), nil)
	if err != nil {
		return err
	}

	resp, err := s.client.doRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	return nil
}

// unsubscribe sends a POST request to unsubscribe from market data.
func (s *MarketDataService) unsubscribe(ctx context.Context, body map[string]interface{}) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.client.BuildURL("/v1/api/iserver/marketdata/unsubscribe"), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	var httpResp *http.Response
	if err := func() error {
		httpResp, err = s.client.doRequest(req)
		return err
	}(); err != nil {
		return err
	}
	defer httpResp.Body.Close()

	return nil
}

// RegulatorySnapshot retrieves regulatory snapshot for a contract.
// WARNING: Each regulatory snapshot request incurs a fee of $0.01 USD.
func (s *MarketDataService) RegulatorySnapshot(ctx context.Context, conid int) (map[string]interface{}, error) {
	params := url.Values{}
	params.Set("conid", strconv.Itoa(conid))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.BuildURL("/v1/api/md/regsnapshot?"+params.Encode()), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result, nil
}

// Helper functions.

func joinInts(strs []string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += ","
		}
		result += s
	}
	return result
}

func joinStrings(strs []string) string {
	result := ""
	for i, s := range strs {
		if i > 0 {
			result += ","
		}
		result += s
	}
	return result
}

// ScannerService handles market scanner API calls.
type ScannerService struct {
	client *Client
}

// ScannerParams represents the response from /iserver/scanner/params.
type ScannerParams struct {
	ScanTypeList   []ScanType   `json:"scan_type_list"`
	InstrumentList []Instrument `json:"instrument_list"`
	FilterList     []Filter     `json:"filter_list"`
	LocationTree   []Location   `json:"location_tree"`
}

// ScanType represents a scanner type.
type ScanType struct {
	DisplayName string   `json:"display_name"`
	Code        string   `json:"code"`
	Instruments []string `json:"instruments"`
}

// Instrument represents an instrument type for scanning.
type Instrument struct {
	DisplayName string   `json:"display_name"`
	Type        string   `json:"type"`
	Filters     []string `json:"filters"`
}

// Filter represents a filter option for scanning.
type Filter struct {
	Group       string `json:"group"`
	DisplayName string `json:"display_name"`
	Code        string `json:"code"`
	Type        string `json:"type"`
}

// Location represents a location for scanning.
type Location struct {
	DisplayName string     `json:"display_name"`
	Type        string     `json:"type"`
	Locations   []Location `json:"locations"`
}

// ScannerRunRequest represents a scanner run request.
type ScannerRunRequest struct {
	Instrument string            `json:"instrument"`
	Location   string            `json:"location"`
	Type       string            `json:"type"`
	Filter     []ScannerFilter   `json:"filter,omitempty"`
}

// ScannerFilter represents a filter for scanner requests.
type ScannerFilter struct {
	Code  string      `json:"code"`
	Value interface{} `json:"value"`
}

// ScannerContract represents a contract found by scanner.
type ScannerContract struct {
	ServerID            string `json:"server_id"`
	ColumnName         string `json:"column_name,omitempty"`
	Symbol             string `json:"symbol"`
	ConidEx            string `json:"conidex"`
	ConID              int    `json:"con_id"`
	AvailableChartPeriods string `json:"available_chart_periods,omitempty"`
	CompanyName        string `json:"company_name"`
	ScanData           string `json:"scan_data,omitempty"`
	ContractDescription string `json:"contract_description_1"`
	ListingExchange    string `json:"listing_exchange"`
	SecType            string `json:"sec_type"`
}

// ScannerResponse represents the response from /iserver/scanner/run.
type ScannerResponse struct {
	Contracts           []ScannerContract `json:"contracts"`
	ScanDataColumnName string            `json:"scan_data_column_name,omitempty"`
}

// Params returns available scanner parameters.
func (s *ScannerService) Params(ctx context.Context) (*ScannerParams, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, s.client.BuildURL("/v1/api/iserver/scanner/params"), nil)
	if err != nil {
		return nil, err
	}

	resp, err := s.client.doRequest(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ScannerParams
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

// Run executes a market scanner with the given parameters.
func (s *ScannerService) Run(ctx context.Context, req ScannerRunRequest) (*ScannerResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, s.client.BuildURL("/v1/api/iserver/scanner/run"), bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := s.client.doRequest(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var result ScannerResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}