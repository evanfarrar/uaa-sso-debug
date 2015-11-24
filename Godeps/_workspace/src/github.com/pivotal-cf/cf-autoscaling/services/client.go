package services

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"net/http"
	"sync"

	"github.com/evanfarrar/uaa-sso-debug/config"
)

var _client *http.Client
var mutex sync.Mutex

func GetClient() *http.Client {
	mutex.Lock()
	defer mutex.Unlock()

	if _client == nil {
		env := config.NewEnvironment()
		_client = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{
					InsecureSkipVerify: !env.VerifySSL,
				},
			},
		}
	}

	return _client
}

type Client struct {
	Host              string
	EnableBasicAuth   bool
	BasicAuthUsername string
	BasicAuthPassword string
	ContentType       string
}

func (c Client) MakeRequest(method, path string, accessToken string, requestBody io.Reader) (int, []byte, error) {
	url := c.Host + path
	request, err := http.NewRequest(method, url, requestBody)
	if err != nil {
		return 0, nil, err
	}
	if c.EnableBasicAuth {
		request.SetBasicAuth(c.BasicAuthUsername, c.BasicAuthPassword)
	}
	if accessToken != "" {
		request.Header.Set("Authorization", "Bearer "+accessToken)
	}
	if requestBody != nil {
		request.Header.Set("Content-Type", c.ContentType)
	}
	client := GetClient()
	response, err := client.Do(request)
	if err != nil {
		return 0, nil, err
	}

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		return response.StatusCode, body, err
	}

	return response.StatusCode, body, nil
}
