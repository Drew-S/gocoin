package main

import (
    "fmt"
    "io/ioutil"
    "os"
    "net/http"
    "encoding/json"
    "strconv"
    "github.com/jessevdk/go-flags"
    "strings"
    "regexp"
    "time"
)

var opts struct {
    Currency string `short:"c" long:"currency" description:"The currency to convert the values to," default:"CAD"`
    Coin string `short:"x" long:"coin" description:"The cryptocurrency to get data." default:"BTC"`
    Format string `short:"f" long:"format" description:"The format of the output string." default:"%C: %P %1D:P %1D:C"`
}

const endpoint string = "https://api.nomics.com/v1/currencies/ticker?key=%s&ids=%s&convert=%s"

type D struct {
    Volume string `json:"volume"`
    PriceChange string `json:"price_change"`
    PriceChangePCT string `json:"price_change_pct"`
    VolumeChange string `json:"volume_change"`
    VolumeChangePCT string `json:"volume_change_pct"`
    MarketCapChange string `json:"market_cap_change"`
    MarketCapChangePCT string `json:"market_cap_change_pct"`
}

type Coin struct {
    Id string `json:"id"`
    Currency string `json:"currency"`
    Symbol string `json:"symbol"`
    Name string `json:"name"`
    LogoURL string `json:"logo_url"`
    Price string `json:"price"`
    PriceDate string `json:"price_date"`
    PriceTimestamp string `json:"price_timestamp"`
    CirculatingSupply string `json:"circulating_supply"`
    MaxSupply string `json:"max_supply"`
    MarketCap string `json:"market_cap"`
    Rank string `json:"rank"`
    High string `json:"high"`
    HighTimestamp string `json:"high_timestamp"`
    D1 D `json:"1d"`
    D7 D `json:"7d"`
    D30 D `json:"30d"`
    D365 D `json:"365d"`
    YTD D `json:"ytd"`
}

const UsageFormat string = ` [Options]

  Wraps the https://api.nomics.com/v1/currencies/ticker cryptocurrency
  open API in a simple terminal interface, for use in Conky and other terminal
  applications. Specify a cryptocurrency to obtain, a currency to convert to
  and the output format.

Formatting Options:
Default:   "%C: %P %1D:P %1D:C" -> "BTC: 15627.42669435 279.96446090 "
  
  %I  id  "BTC"
  %C  currency  "BTC"
  %S  symbol  "BTC"
  %N  name  "Bitcoin"
  %L  logo_url  "https://s3.us-east-2..."
  %P  price  "11616.76734947"
  %D  price_date  "2020-08-30T00..."
  %T  price_timestamp  "2020-08-30..."
  %CS  circulatingSupply  "18475000"
  %M  max_supply  "21000000"
  %MC  market_cap  "214619776781"
  %R  rank  "1"
  %H  high  "19337.69352527"
  %HT  high_timestamp  "2017-12-16..."
  %$  Currency being displayed in "CAD"
  %1D:, %7D:, %30D:, %365D:, %YTD:  selections for each volume date which each have
      sub options.
  
      %1D:V  volume  "17087256040.52"
      %1D:P  price_change  "116.36..."
      %1D:PP  price_change_pct  "0.0101"
      %1D:VC  volume_change  "-48633..."
      %1D:VP  volume_change_pct  "-0.0277"
      %1D:M  market_cap_change  "21590..."
      %1D:MP  market_cap_change_pct  "0.0102"
      %1D:C   price_change_carret  "" or ""
  
      Format for sub options: %1D: 05V
  
  %% Just a normal percent sign "%"
  
  Any string or integer format value can contain padding control.
      e.g. %4C -> ' BTC',
           %-4C -> 'BTC '
           %03R -> '001',
           % 4H -> ' 19337.69352527' or '-19337.69352527'
  
  Any floating point format value can contain both padding controland decimal
  precision control.
      e.g. %0.3H -> '19337.693',
  
  Formatting time strings (%D, %T, and %HT) is done using golangs time formatting
  https://golang.org/pkg/time/ in between %{<format>}D
      e.g. %{Mon Jan _2 2006}D -> 'Tue Sep  1 2020'`

// Parses a time value component
func parseTime(format, pre, post, value string) string {
    re := regexp.MustCompile(pre + "{(.*)}" + post)
    t, err := time.Parse("2006-01-02T15:04:05Z", value)
    if err != nil {
        return value
    }
    return re.ReplaceAllStringFunc(format, func(val string) string {
        str := re.FindStringSubmatch(val)[1]
        return t.Format(str)
    })
}

// Parses a replacement value interpreting the pad-formatting numbers, uses
// Golang's built in formatting. See https://golang.org/pkg/fmt/
func parsePad(fl bool, format string, pre string, post string, value string) string {
    if fl {
        re := regexp.MustCompile(pre + `(\s?-?\d*\.?\d+)` + post)
        format = re.ReplaceAllStringFunc(format, func(val string) string {
            str := re.FindStringSubmatch(val)[1]
            f, _ := strconv.ParseFloat(value, 64)
            return fmt.Sprintf("%" + str + "f", f)
        })
    } else {
        re := regexp.MustCompile(pre + `(-?\d+)` + post)
        format = re.ReplaceAllStringFunc(format, func(val string) string {
            str := re.FindStringSubmatch(val)[1]
            return fmt.Sprintf("%" + str + "s", value)
        })
    }
    return strings.ReplaceAll(format, pre + post, value)
}

// Parses all the options for the root option of the currency
func parse(coin Coin, format string, currency string) string {
    format = parseTime(format, "%", "HT", coin.HighTimestamp)
    format = parsePad(false, format, "%", "MC", coin.MarketCap)
    format = parsePad(false, format, "%", "CS", coin.CirculatingSupply)
    format = parsePad(true, format, "%", "H", coin.High)
    format = parsePad(false, format, "%", "M", coin.MaxSupply)
    format = parseTime(format, "%", "T", coin.PriceTimestamp)
    format = parseTime(format, "%", "D", coin.PriceDate)
    format = parsePad(true, format, "%", "P", coin.Price)
    format = parsePad(false, format, "%", "L", coin.LogoURL)
    format = parsePad(false, format, "%", "N", coin.Name)
    format = parsePad(false, format, "%", "s", coin.Symbol)
    format = parsePad(false, format, "%", "C", coin.Currency)
    format = parsePad(false, format, "%", "I", coin.Id)
    format = parsePad(false, format, "%", "$", currency)
    return format
}

// Parses all the options for a historical period of the currency
func parseDay(pre string, day D, format string) string {
    format = parsePad(true, format, pre, "PP", day.PriceChangePCT)
    format = parsePad(true, format, pre, "VC", day.VolumeChange)
    format = parsePad(true, format, pre, "VP", day.VolumeChangePCT)
    format = parsePad(true, format, pre, "MP", day.MarketCapChangePCT)
    format = parsePad(true, format, pre, "V", day.Volume)
    format = parsePad(true, format, pre, "P", day.PriceChange)
    format = parsePad(true, format, pre, "M", day.MarketCapChange)
    var good string = ""
    if val, _ := strconv.ParseFloat(day.PriceChange, 64); val < 0 {
        good = ""
    }
    format = parsePad(false, format, pre, "C", good)
    return format
}

// Run the program obtaining coin data
func main() {
    var api string
    var currency string = "CAD"
    var coin string = "BTC"

    var format string = "%C: %P %1D:P %1D:C\n"

    // Get the API key token
    var dir string = "~/.gocoin"

    if os.Getenv("XDG_CONFIG_HOME") != "" {
        dir = os.Getenv("XDG_CONFIG_HOME") + "/gocoin"
    }

    key, err := ioutil.ReadFile(dir + "/key")
    if err != nil {
        fmt.Printf("An error occured: %s\n", err.Error())
    }

    // Clean key
    api = strings.TrimRight(string(key), "\n")

    // Parse options
    parser := flags.NewParser(&opts, flags.Default)
    parser.Usage = UsageFormat
    if _, err  = parser.Parse(); err != nil {
        return
    }

    // Set local currency (CAD, USD, EUR, etc.)
    if opts.Currency != "" {
        currency = opts.Currency
    }
    // Set target cryptocurrency (BTC, BAT, ETH, etc.)
    if opts.Coin != "" {
        coin = opts.Coin
    }
    // Sets the formatting string
    if opts.Format != "" {
        format = opts.Format
    }

    // Get remote data from nomics.com
    response, err := http.Get(fmt.Sprintf(endpoint, api, coin, currency))
    if err != nil {
        fmt.Printf("An error occured: %s\n", err.Error())
    }

    defer response.Body.Close()

    c := make([]Coin, 1) 

    responseData, err := ioutil.ReadAll(response.Body)
    if err != nil {
        fmt.Printf("An error occured: %s\n", err.Error())
    }

    json.Unmarshal(responseData, &c)

    re := regexp.MustCompile("\\\\(.)")

    // Format and print the response data
    format = strings.ReplaceAll(format, "%%", "::{PERCENT}::")
    format = parseDay("%1D:", c[0].D1, format)
    format = parseDay("%7D:", c[0].D7, format)
    format = parseDay("%30D:", c[0].D30, format)
    format = parseDay("%365D:", c[0].D365, format)
    format = parseDay("%YTD:", c[0].YTD, format)
    format = parse(c[0], format, currency)
    format = strings.ReplaceAll(format, "::{PERCENT}::", "%")
    format = re.ReplaceAllStringFunc(format, func(val string) string {
        return re.FindStringSubmatch(val)[1]     
    })

    fmt.Println(format)
}
