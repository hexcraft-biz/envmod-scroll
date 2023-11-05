package scroll

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	"github.com/hexcraft-biz/her"
	"github.com/hexcraft-biz/misc/xtime"
)

// ================================================================
// Short link service
// ================================================================
type Scroll struct {
	Host                string
	EndpointGetShortUrl *url.URL
}

func New() (*Scroll, *her.Error) {
	u, err := url.ParseRequestURI(os.Getenv("HOST_SCROLL"))
	if err != nil {
		return nil, her.NewError(http.StatusInternalServerError, err, nil)
	}

	return &Scroll{
		Host:                u.String(),
		EndpointGetShortUrl: u.JoinPath("/links/v1/urls"),
	}, nil
}

// ================================================================
//
// ================================================================
const (
	MinDuration = 600
)

type inputGetShortUrl struct {
	RedirectUri string      `json:"redirectUri"`
	StartedAt   *xtime.Time `json:"startedAt,omitempty"`
	Duration    *int        `json:"duration,omitempty"`
}

type Request struct {
	*http.Request
}

func (e Scroll) NewRequestGetShortUrl(redirectUri *url.URL, startedAt *xtime.Time, duration int) (*Request, *her.Error) {
	input := &inputGetShortUrl{
		RedirectUri: redirectUri.String(),
		StartedAt:   startedAt,
	}

	if duration > MinDuration {
		input.Duration = &duration
	}

	jsonbytes, err := json.Marshal(input)
	if err != nil {
		return nil, her.NewError(http.StatusInternalServerError, err, nil)
	}

	req, err := http.NewRequest("POST", e.EndpointGetShortUrl.String(), bytes.NewReader(jsonbytes))
	if err != nil {
		return nil, her.NewError(http.StatusInternalServerError, err, nil)
	}

	return &Request{Request: req}, nil
}

func (r Request) Do() (string, *her.Error) {
	client := &http.Client{}
	result := new(struct {
		Url string `json:"url"`
	})
	payload := her.NewPayload(result)
	if resp, err := client.Do(r.Request); err != nil {
		return "", her.NewError(http.StatusInternalServerError, err, nil)
	} else if err := her.FetchHexcApiResult(resp, payload); err != nil {
		return "", err
	} else if resp.StatusCode >= 400 {
		return "", her.NewErrorWithMessage(http.StatusInternalServerError, payload.Message, nil)
	} else {
		return result.Url, nil
	}
}
