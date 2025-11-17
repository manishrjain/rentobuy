package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
)

const marketDataFile = ".rentobuy_market_data.json"

// MarketData stores historical annual returns
type MarketData struct {
	LastUpdated string             `json:"last_updated"`
	SP500       map[string]float64 `json:"sp500"`  // Year -> Annual return %
	QQQ         map[string]float64 `json:"qqq"`    // Year -> Annual return %
}

// YahooChartResponse represents the JSON response from Yahoo Finance chart API
type YahooChartResponse struct {
	Chart struct {
		Result []struct {
			Timestamp []int64 `json:"timestamp"`
			Indicators struct {
				Adjclose []struct {
					Adjclose []float64 `json:"adjclose"`
				} `json:"adjclose"`
			} `json:"indicators"`
		} `json:"result"`
	} `json:"chart"`
}

// fetchYahooFinanceData fetches historical price data from Yahoo Finance using chart API
func fetchYahooFinanceData(ticker string, startDate, endDate time.Time) ([][]string, error) {
	// Convert to Unix timestamps
	period1 := startDate.Unix()
	period2 := endDate.Unix()

	// Build URL using chart API (more reliable than download endpoint)
	url := fmt.Sprintf("https://query2.finance.yahoo.com/v8/finance/chart/%s?period1=%d&period2=%d&interval=1d",
		ticker, period1, period2)

	// Create request with headers
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_15_7) AppleWebKit/537.36")

	// Make request
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch data: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("yahoo finance returned status %d", resp.StatusCode)
	}

	// Parse JSON
	var chartResp YahooChartResponse
	err = json.NewDecoder(resp.Body).Decode(&chartResp)
	if err != nil {
		return nil, fmt.Errorf("failed to parse JSON: %v", err)
	}

	if len(chartResp.Chart.Result) == 0 {
		return nil, fmt.Errorf("no data returned")
	}

	result := chartResp.Chart.Result[0]
	timestamps := result.Timestamp
	adjCloses := result.Indicators.Adjclose[0].Adjclose

	if len(timestamps) != len(adjCloses) {
		return nil, fmt.Errorf("data length mismatch")
	}

	// Convert to CSV format: Date, Adj Close
	records := [][]string{{"Date", "Adj Close"}}
	for i, ts := range timestamps {
		date := time.Unix(ts, 0).Format("2006-01-02")
		adjClose := fmt.Sprintf("%.6f", adjCloses[i])
		records = append(records, []string{date, adjClose})
	}

	return records, nil
}

// calculateAnnualReturns calculates annual returns from daily price data
func calculateAnnualReturns(records [][]string) (map[string]float64, error) {
	if len(records) < 2 {
		return nil, fmt.Errorf("insufficient data")
	}

	// Skip header row
	records = records[1:]

	// Group by year and get first/last prices
	type yearData struct {
		firstPrice float64
		lastPrice  float64
	}
	yearPrices := make(map[string]*yearData)

	for _, record := range records {
		if len(record) < 2 {
			continue
		}

		// Parse date (format: YYYY-MM-DD)
		date := record[0]
		if len(date) < 4 {
			continue
		}
		year := date[:4]

		// Parse adjusted close price (column 1)
		adjClose, err := strconv.ParseFloat(record[1], 64)
		if err != nil {
			continue
		}

		// Initialize year data if needed
		if yearPrices[year] == nil {
			yearPrices[year] = &yearData{firstPrice: adjClose, lastPrice: adjClose}
		}

		// Update last price (data is in chronological order)
		yearPrices[year].lastPrice = adjClose
	}

	// Calculate annual returns
	returns := make(map[string]float64)
	for year, data := range yearPrices {
		if data.firstPrice > 0 {
			returnPct := ((data.lastPrice - data.firstPrice) / data.firstPrice) * 100
			returns[year] = returnPct
		}
	}

	return returns, nil
}

// loadMarketData loads cached market data from file
func loadMarketData() (*MarketData, error) {
	data, err := os.ReadFile(marketDataFile)
	if err != nil {
		if os.IsNotExist(err) {
			return &MarketData{
				SP500: make(map[string]float64),
				QQQ:   make(map[string]float64),
			}, nil
		}
		return nil, err
	}

	var md MarketData
	err = json.Unmarshal(data, &md)
	if err != nil {
		return nil, err
	}

	if md.SP500 == nil {
		md.SP500 = make(map[string]float64)
	}
	if md.QQQ == nil {
		md.QQQ = make(map[string]float64)
	}

	return &md, nil
}

// saveMarketData saves market data to cache file
func saveMarketData(md *MarketData) error {
	md.LastUpdated = time.Now().Format("2006-01-02")

	data, err := json.MarshalIndent(md, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(marketDataFile, data, 0644)
}

// updateMarketData fetches and updates market data if needed
func updateMarketData() (*MarketData, error) {
	md, err := loadMarketData()
	if err != nil {
		return nil, fmt.Errorf("failed to load cache: %v", err)
	}

	// Check if we need to update
	now := time.Now()
	needsUpdate := false

	// Update if cache is older than 1 day
	if md.LastUpdated != "" {
		lastUpdate, err := time.Parse("2006-01-02", md.LastUpdated)
		if err == nil {
			if now.Sub(lastUpdate) > 24*time.Hour {
				needsUpdate = true
			}
		}
	} else {
		needsUpdate = true
	}

	// Also update if we don't have current year data
	currentYear := fmt.Sprintf("%d", now.Year())
	if _, ok := md.SP500[currentYear]; !ok {
		needsUpdate = true
	}

	if !needsUpdate {
		return md, nil
	}

	fmt.Println("Updating market data from Yahoo Finance...")

	// Fetch data for last 11 years (to ensure we have complete 10 years)
	startDate := time.Now().AddDate(-11, 0, 0)
	endDate := time.Now()

	// Fetch S&P 500
	sp500Records, err := fetchYahooFinanceData("^GSPC", startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch S&P 500 data: %v", err)
	}

	sp500Returns, err := calculateAnnualReturns(sp500Records)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate S&P 500 returns: %v", err)
	}

	// Wait a bit to avoid rate limiting
	time.Sleep(1 * time.Second)

	// Fetch QQQ
	qqqRecords, err := fetchYahooFinanceData("QQQ", startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch QQQ data: %v", err)
	}

	qqqReturns, err := calculateAnnualReturns(qqqRecords)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate QQQ returns: %v", err)
	}

	// Update cache with new data
	for year, ret := range sp500Returns {
		md.SP500[year] = ret
	}
	for year, ret := range qqqReturns {
		md.QQQ[year] = ret
	}

	// Save to cache
	err = saveMarketData(md)
	if err != nil {
		return nil, fmt.Errorf("failed to save cache: %v", err)
	}

	fmt.Println("Market data updated successfully.")

	return md, nil
}

// calculateMarketAverages calculates 10-year averages for S&P 500 and QQQ
func calculateMarketAverages(md *MarketData) (float64, float64) {
	if md == nil {
		return 0, 0
	}

	var sp500Sum, qqqSum float64
	count := 0

	currentYear := time.Now().Year()

	for year, sp500Ret := range md.SP500 {
		yearInt, _ := strconv.Atoi(year)
		// Only include complete years (not current year) from last 10 years
		if yearInt >= currentYear-10 && yearInt < currentYear {
			if qqqRet, ok := md.QQQ[year]; ok {
				sp500Sum += sp500Ret
				qqqSum += qqqRet
				count++
			}
		}
	}

	if count == 0 {
		return 0, 0
	}

	return sp500Sum / float64(count), qqqSum / float64(count)
}

// displayMarketData shows historical returns and averages
func displayMarketData(md *MarketData) {
	fmt.Println("\n=== MARKET DATA ===")

	// Get sorted years
	years := make([]string, 0)
	for year := range md.SP500 {
		// Only show last 10 complete years
		yearInt, _ := strconv.Atoi(year)
		if yearInt >= time.Now().Year()-10 {
			years = append(years, year)
		}
	}
	sort.Strings(years)

	// Display table
	fmt.Printf("\n%-15s %-18s %-18s\n", "Period", "S&P 500", "QQQ")
	fmt.Println(strings.Repeat("-", 55))

	var sp500Sum, qqqSum float64
	count := 0

	for _, year := range years {
		sp500Ret := md.SP500[year]
		qqqRet := md.QQQ[year]

		// Only include in average if it's a complete year (not current year)
		if year != fmt.Sprintf("%d", time.Now().Year()) {
			sp500Sum += sp500Ret
			qqqSum += qqqRet
			count++
		}

		fmt.Printf("MRKT   %-8s %-18s %-18s\n", year,
			fmt.Sprintf("%.2f%%", sp500Ret),
			fmt.Sprintf("%.2f%%", qqqRet))
	}

	if count > 0 {
		fmt.Println(strings.Repeat("-", 55))
		fmt.Printf("MRKT   %-8s %-18s %-18s\n", "Average",
			fmt.Sprintf("%.2f%%", sp500Sum/float64(count)),
			fmt.Sprintf("%.2f%%", qqqSum/float64(count)))
	}
}
