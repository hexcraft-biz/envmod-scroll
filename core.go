package scroll

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/url"
	"os"

	resph "github.com/hexcraft-biz/misc/resp"
	"github.com/hexcraft-biz/misc/xtime"
)

// ================================================================
// Short link service
// ================================================================
type Scroll struct {
	Host                string
	EndpointGetShortUrl *url.URL
}

func New() (*Scroll, *resph.Resp) {
	u, err := url.ParseRequestURI(os.Getenv("HOST_SCROLL"))
	if err != nil {
		return nil, resph.NewError(http.StatusInternalServerError, err, nil)
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

func (e Scroll) NewRequestGetShortUrl(redirectUri *url.URL, startedAt *xtime.Time, duration int) (*Request, *resph.Resp) {
	input := &inputGetShortUrl{
		RedirectUri: redirectUri.String(),
		StartedAt:   startedAt,
	}

	if duration > MinDuration {
		input.Duration = &duration
	}

	jsonbytes, err := json.Marshal(input)
	if err != nil {
		return nil, resph.NewError(http.StatusInternalServerError, err, nil)
	}

	req, err := http.NewRequest("POST", e.EndpointGetShortUrl.String(), bytes.NewReader(jsonbytes))
	if err != nil {
		return nil, resph.NewError(http.StatusInternalServerError, err, nil)
	}

	return &Request{Request: req}, nil
}

func (r Request) Do() (string, *resph.Resp) {
	client := &http.Client{}
	result := new(struct {
		Url string `json:"url"`
	})
	payload := resph.NewPayload(result)
	if resp, err := client.Do(r.Request); err != nil {
		return "", resph.NewError(http.StatusInternalServerError, err, nil)
	} else if err := resph.FetchHexcApiResult(resp, payload); err != nil {
		return "", err
	} else if resp.StatusCode >= 400 {
		return "", resph.NewErrorWithMessage(http.StatusInternalServerError, payload.Message, nil)
	} else {
		return result.Url, nil
	}
}
