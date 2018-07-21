package favqs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"
)

const (
	baseURL = "https://favqs.com/api/"
)

var (
	// ErrAPIKey indicates that the FAVQS_APIKEY needed to hit the API is missing
	ErrAPIKey = fmt.Errorf("FAVQS_APIKEY is missing from the environment")
	// ErrNoQuotes ...
	ErrNoQuotes = fmt.Errorf("No quotes found")
	// ErrStatusNotOK ...
	ErrStatusNotOK = fmt.Errorf("HTTP status was not 200")
	// ErrSessionFailed ...
	ErrSessionFailed = fmt.Errorf("Session creation failed")
)

var (
	// DefaultFilters provides a easy defualt filter for my use case
	DefaultFilters = []string{"beauty", "inspirational", "art"}
	// DefaultMax provides the number of quotes to return for my use case
	DefaultMax = 10
)

// QuoteOfTheDate response structure from favq
type QuoteOfTheDate struct {
	QotdDate string `json:"qotd_date"`
	Quote    Quote  `json:"quote"`
}

// Quotes returns a paged list of quotes from favq
type Quotes struct {
	Page     int     `json:"page"`
	LastPage bool    `json:"last_page"`
	Quotes   []Quote `json:"quotes"`
}

// Quote base structure
type Quote struct {
	ID              int      `json:"id,omitempty"`
	FavoritesCount  int      `json:"favorites_count,omitempty"`
	Dialogue        bool     `json:"dialogue,omitempty"`
	Favorite        bool     `json:"favorite,omitempty"`
	Tags            []string `json:"tags,omitempty"`
	URL             string   `json:"url,omitempty"`
	UpvotesCount    int      `json:"upvotes_count,omitempty"`
	DownvotesCount  int      `json:"downvotes_count,omitempty"`
	Author          string   `json:"author,omitempty"`
	AuthorPermalink string   `json:"author_permalink,omitempty"`
	Body            string   `json:"body,omitempty"`
}

// Client a wrapper favqs
type Client struct {
	httpClient  *http.Client
	authHeaders http.Header
}

// New returns a Client ready to interact with favsq
// your APIKEY is required
func New() (Client, error) {
	var c Client
	key := strings.TrimSpace(os.Getenv("FAVQS_APIKEY"))
	if len(key) == 0 {
		return c, ErrAPIKey
	}

	cj, err := cookiejar.New(nil)
	if err != nil {
		return c, err
	}

	client := &http.Client{
		Transport: &http.Transport{
			Proxy: http.ProxyFromEnvironment,
			Dial: (&net.Dialer{
				Timeout: 5 * time.Second,
			}).Dial,
			TLSHandshakeTimeout: 5 * time.Second,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
		},
		Timeout: 5 * time.Second,
		Jar:     cj,
	}

	h := make(http.Header)
	h.Set("Authorization", "Token token="+key)
	authHeaders := h
	c.authHeaders = authHeaders
	c.httpClient = client

	return c, nil
}

// GetQuoteOfTheDay ...
func (c Client) GetQuoteOfTheDay() (QuoteOfTheDate, error) {
	var quote QuoteOfTheDate
	req, err := http.NewRequest("GET", baseURL+"qotd", nil)
	if err != nil {
		return quote, err
	}
	// not required, but might has well use it since it's already present
	req.Header = c.authHeaders

	res, err := c.httpClient.Do(req)
	if err != nil {
		return quote, err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return quote, err
	}

	err = json.Unmarshal(b, &quote)
	if err != nil {
		return quote, err
	}

	return quote, nil
}

// GetQuotes returns random quotes from filter, defaults to tag type
func (c Client) GetQuotes(filter string, max int) ([]Quote, error) {
	u, err := url.Parse(baseURL + "quotes")
	if err != nil {
		return nil, err
	}
	params := u.Query()
	params.Add("filter", filter)
	params.Add("type", "tag")
	u.RawQuery = params.Encode()
	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header = c.authHeaders

	res, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var tqs Quotes
	err = json.Unmarshal(b, &tqs)
	if err != nil {
		return nil, err
	}

	if len(tqs.Quotes) == 0 {
		return nil, ErrNoQuotes
	}

	if len(tqs.Quotes) < max {
		max = len(tqs.Quotes)
	}

	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	indexes := r.Perm(max)

	quotes := make([]Quote, max)
	for i, item := range indexes {
		if i < len(tqs.Quotes) {
			quotes[i] = tqs.Quotes[item]
		}
	}

	return quotes, nil
}

// GetRandomFilterFromDefaults returns a random filter from the default list
func GetRandomFilterFromDefaults() string {
	s := rand.NewSource(time.Now().UnixNano())
	r := rand.New(s)
	i := r.Perm(len(DefaultFilters))[0]
	return DefaultFilters[i]
}
