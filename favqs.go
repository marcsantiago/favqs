package favqs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/pkg/errors"
)

const (
	baseURL = "https://favqs.com/api/"
)

var (
	// ErrAPIKey indicates that the APIKEY needed to hit the API is missing
	ErrAPIKey = fmt.Errorf("APIKEY is missing from the environment")
	// ErrStatusNotOK ...
	ErrStatusNotOK = fmt.Errorf("HTTP status was not 200")
	// ErrSessionFailed ...
	ErrSessionFailed = fmt.Errorf("Session creation failed")
)

var (
	// DefaultFilter provides a easy defualt filter for my use case
	DefaultFilter = "fashion"
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

// New returns a fully sessioned Client to interact with favsq
// your APIKEY is required
func New() (Client, error) {
	var c Client
	key := strings.TrimSpace(os.Getenv("APIKEY"))
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

	req, err := http.NewRequest("POST", baseURL+"session", nil)
	if err != nil {
		return c, err
	}
	req.Header.Set("Authorization", "Token token="+key)

	res, err := client.Do(req)
	if err != nil {
		return c, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return c, errors.Wrapf(ErrStatusNotOK, "wanted 200 code, got: %d", res.StatusCode)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return c, err
	}

	responseWrapper := struct {
		UserToken string `json:"User-Token,omitempty"`
		ErrorCode string `json:"error_code"`
		Message   string `json:"message"`
	}{}

	err = json.Unmarshal(b, &responseWrapper)
	if err != nil {
		return c, err
	}

	if len(responseWrapper.ErrorCode) > 0 {
		return c, errors.Wrapf(ErrSessionFailed, "got error_code %s with the message %s", responseWrapper.ErrorCode, responseWrapper.Message)
	}
	req.Header.Set("User-Token", responseWrapper.UserToken)

	authHeaders := req.Header
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

// GetQuotes returns quotes from filter, defaults to tag type
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

	quotes := make([]Quote, max)
	for i, q := range tqs.Quotes {
		if i < len(tqs.Quotes) {
			quotes[i] = q
		}
	}

	return quotes, nil
}
