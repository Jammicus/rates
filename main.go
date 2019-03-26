package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"strings"

	"github.com/sirupsen/logrus"
	log "github.com/sirupsen/logrus"
)

type Response interface {
	printInfo()
	parseRequest(http.Response) (Response, error)
}

var Api = "https://api.exchangeratesapi.io/"
var BaseFlag, StartFlag, EndFlag, CurrencyFlag *string

type ResponseLatest struct {
	Base  string             `json:"base"`
	Rates map[string]float64 `json:"rates"`
	Date  string             `json:"date"`
}

type ResponseHistory struct {
	Base  string           `json:"base"`
	Rates map[string]Rates `json:"rates"`
}

type Rates struct {
	IDR float64 `json:"IDR"`
	DKK float64 `json:"DDK"`
	INR float64 `json:"INR"`
	HRK float64 `json:"HRK"`
	KRW float64 `json:"KRW"`
	RUB float64 `json:"RUB"`
	ZAR float64 `json:"ZAR"`
	HUF float64 `json:"HUF"`
	MXN float64 `json:"MXN"`
	ISK float64 `json:"ISK"`
	CNY float64 `json:"CNY"`
	USD float64 `json:"USD"`
	TRY float64 `json:"TRY"`
	CZK float64 `json:"CZK"`
	ILS float64 `json:"ILS"`
	JPY float64 `json:"JPY"`
	AUD float64 `json:"AUD"`
	MYR float64 `json:"MYR"`
	BRL float64 `json:"BRL"`
	RON float64 `json:"RON"`
	PHP float64 `json:"PHP"`
	CHF float64 `json:"CHF"`
	SGD float64 `json:"SGD"`
	BGN float64 `json:"BGN"`
	NZD float64 `json:"NZD"`
	THB float64 `json:"THB"`
	NOK float64 `json:"NOK"`
	GBP float64 `json:"GBP"`
	PLN float64 `json:"PLN"`
	SEK float64 `json:"SEK"`
	CAD float64 `json:"CAD"`
	HKD float64 `json:"HKD"`
}

func main() {
	setLogging()
	cmd := flag.Arg(0)

	if cmd == "latest" {
		var r ResponseLatest
		req, err := generateRequest(cmd, *BaseFlag, *StartFlag, *EndFlag, *CurrencyFlag)
		if err != nil {
			log.Fatal(err)
		}
		httpReq, err := sendRequest(req)
		if err != nil {
			log.Fatal(err)
		}
		output, err := r.parseRequest(*httpReq)
		if err != nil {
			log.Fatal(err)
		}
		printResponce(output)
	}

	if cmd == "history" {
		var r ResponseHistory
		req, err := generateRequest(cmd, *BaseFlag, *StartFlag, *EndFlag, *CurrencyFlag)
		if err != nil {
			log.Fatal(err)
		}
		httpReq, err := sendRequest(req)
		if err != nil {
			log.Fatal(err)
		}
		output, err := r.parseRequest(*httpReq)
		if err != nil {
			log.Fatal(err)
		}
		printResponce(output)
	}
}

// Think about how to remove reflection
func (r ResponseHistory) printInfo() {
	fmt.Println("Base Currency:", r.Base)

	for k, v := range r.Rates {
		fmt.Println("")
		fmt.Println("Date:", k)
		fmt.Println("")
		elem := reflect.ValueOf(&v).Elem()

		for i := 0; i < elem.NumField(); i++ {
			fmt.Printf("Currency %s = %v\n",
				elem.Type().Field(i).Name, elem.Field(i).Interface())
		}
	}
}

func (r ResponseLatest) printInfo() {
	fmt.Println("Base Currency:", r.Base)
	fmt.Println("")
	fmt.Println("Date:", r.Date)
	fmt.Println("")
	for k, v := range r.Rates {
		fmt.Printf("Currency %s = %f\n", k, v)
	}
}

func printResponce(r Response) {
	r.printInfo()
}

func setLogging() {
	customFormatter := new(logrus.TextFormatter)
	customFormatter.TimestampFormat = "2006-01-02 15:04:05"
	logrus.SetFormatter(customFormatter)
	customFormatter.FullTimestamp = true
}

func usage() {
	log.Error("Invalid flag, please use one of the following:")
	flag.PrintDefaults()
	os.Exit(2)
}

// Has to be flags then command, due to rules of the flag pakage.
func init() {
	BaseFlag = flag.String("base", "", "Specifies the base currency to use")
	StartFlag = flag.String("start", "", "Specifies the start date to use for a time series")
	EndFlag = flag.String("end", "", "Specifies the end date to use for a time series")
	CurrencyFlag = flag.String("currency", "", "Specifies a comma seperated list of currencies to be used")
	flag.Usage = usage
	flag.Parse()
}

func generateRequest(command, base, start, end, currency string) (string, error) {

	if command != "latest" && command != "history" {
		return "", errors.New("Invalid command " + command)
	}

	if start != "" && end == "" {
		return "", errors.New("Please provide the end flag when doing a time query")
	}
	if end != "" && start == "" {
		return "", errors.New("Please provide the start flag when doing a time query")
	}

	if command == "latest" && base == "" && currency == "" {
		return Api + command, nil
	}

	if command == "latest" && base != "" && currency == "" {
		return Api + command + "?" + "base=" + strings.ToUpper(base), nil
	}

	if command == "latest" && base == "" && currency != "" {
		return Api + command + "?" + "symbols=" + strings.ToUpper(currency), nil
	}

	if command == "latest" && base != "" && currency != "" {
		return Api + command + "?" + "symbols=" + strings.ToUpper(currency) + "&base=" + strings.ToUpper(base), nil
	}

	if command == "history" && start != "" && end != "" && base == "" && currency == "" {
		return Api + command + "?" + "start_at=" + start + "&end_at=" + end, nil
	}

	if command == "history" && start != "" && end != "" && base != "" && currency == "" {
		return Api + command + "?" + "start_at=" + start + "&end_at=" + end + "&" + "base=" + strings.ToUpper(base), nil
	}

	if command == "history" && start != "" && end != "" && base == "" && currency != "" {
		return Api + command + "?" + "start_at=" + start + "&end_at=" + end + "&symbols=" + strings.ToUpper(currency), nil
	}

	if command == "history" && start != "" && end != "" && base != "" && currency != "" {
		return Api + command + "?" + "start_at=" + start + "&end_at=" + end + "&symbols=" + strings.ToUpper(currency) + "&base=" + strings.ToUpper(base), nil
	}

	return "", errors.New("Unexpected request. Please try again")
}

func sendRequest(req string) (*http.Response, error) {
	resp, err := http.Get(req)
	return resp, err
}

func (r ResponseHistory) parseRequest(resp http.Response) (Response, error) {
	var responseHistory ResponseHistory
	var errr error

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return responseHistory, err
	}

	// Improve this.

	errr = json.Unmarshal(responseData, &responseHistory)

	return responseHistory, errr
}

func (r ResponseLatest) parseRequest(resp http.Response) (Response, error) {
	var responseLatest ResponseLatest
	var errr error

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return responseLatest, err
	}

	errr = json.Unmarshal(responseData, &responseLatest)
	return responseLatest, errr
}
