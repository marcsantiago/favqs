package favqs

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net"
	"net/http"
	"net/http/cookiejar"
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
func (c Client) GetQuoteOfTheDay() string {
	return ""
}
