package laya

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultBaseURL = "http://127.0.0.1:8792"
	modelEnglish   = "english"
	Instructions   = "Pick one label."
	requestTimeout = 60 * time.Second
)

// ErrKey is returned before any dial when the API key is empty.
var ErrKey = errors.New("laya api key not set")

// ErrTransport is a dial, timeout, or refused redirect.
var ErrTransport = errors.New("laya transport")

// ErrTruncated is a cut state. It is not a label.
var ErrTruncated = errors.New("laya usage truncated")

// State is the systemone state object.
type State struct {
	Body string `json:"body"`
}

// Question is one choice question.
type Question struct {
	Type         string            `json:"type"`
	Instructions string            `json:"instructions"`
	Criteria     map[string]string `json:"criteria"`
}

// Request is the POST /v1/systemone body.
type Request struct {
	State     State               `json:"state"`
	Questions map[string]Question `json:"questions"`
	Model     string              `json:"model,omitempty"`
}

// Parsed is the verdict answer. A truncated response never becomes a Parsed value.
type Parsed struct {
	Choice           string
	Confidence       float64
	AnswerConfidence float64
	Model            string
}

// Client calls a loopback laya-serve.
type Client struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
}

// NewClient checks the base URL. An empty URL becomes http://127.0.0.1:8792.
// An empty API key is legal here; SystemOne returns ErrKey before dialing.
func NewClient(baseURL, apiKey string) (*Client, error) {
	normalized, err := normalizeBase(baseURL)
	if err != nil {
		return nil, err
	}
	return &Client{
		baseURL: normalized,
		apiKey:  strings.TrimSpace(apiKey),
		httpClient: &http.Client{
			Timeout:       requestTimeout,
			CheckRedirect: refuseRedirect,
			Transport:     loopbackTransport(),
		},
	}, nil
}

func refuseRedirect(req *http.Request, via []*http.Request) error {
	if req != nil {
		req.Header.Del("Authorization")
	}
	return ErrTransport
}

func loopbackTransport() http.RoundTripper {
	base, ok := http.DefaultTransport.(*http.Transport)
	if !ok || base == nil {
		return &http.Transport{Proxy: nil}
	}
	tr := base.Clone()
	tr.Proxy = nil
	return tr
}

func normalizeBase(raw string) (string, error) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		raw = defaultBaseURL
	}
	if err := validateBaseURL(raw); err != nil {
		return "", err
	}
	u, err := url.Parse(raw)
	if err != nil {
		return "", errors.New("laya base url rejected")
	}
	u.Path = ""
	u.RawQuery = ""
	u.Fragment = ""
	return u.String(), nil
}

func validateBaseURL(raw string) error {
	u, err := url.Parse(raw)
	if err != nil || u.Host == "" || u.User != nil {
		return errors.New("laya base url rejected")
	}
	if !bareHTTPURL(u) {
		return errors.New("laya base url rejected")
	}
	ip := net.ParseIP(u.Hostname())
	if ip == nil || !loopbackIP(ip) {
		return errors.New("laya base url rejected")
	}
	return nil
}

func loopbackIP(ip net.IP) bool {
	if ip4 := ip.To4(); ip4 != nil {
		return ip4[0] == 127
	}
	return ip.Equal(net.IPv6loopback)
}

func (c *Client) systemOneURL() (string, error) {
	if err := validateBaseURL(c.baseURL); err != nil {
		return "", err
	}
	return strings.TrimRight(c.baseURL, "/") + "/v1/systemone", nil
}

// SystemOne posts one choice question. model is always english.
func (c *Client) SystemOne(ctx context.Context, body string, criteria map[string]string) (Parsed, error) {
	if strings.TrimSpace(c.apiKey) == "" {
		return Parsed{}, ErrKey
	}
	endpoint, err := c.systemOneURL()
	if err != nil {
		return Parsed{}, err
	}
	req, err := c.newRequest(ctx, endpoint, body, criteria)
	if err != nil {
		return Parsed{}, err
	}
	resp, err := c.httpClient.Do(req)
	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
	}
	if err != nil {
		return Parsed{}, fmt.Errorf("%w: %s", ErrTransport, scrub(err.Error(), c.apiKey, body))
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 2048))
		return Parsed{}, fmt.Errorf("laya http %d: %s", resp.StatusCode, scrub(detailText(b), c.apiKey, body))
	}
	return decodeVerdict(resp.Body)
}

func decodeVerdict(r io.Reader) (Parsed, error) {
	var decoded struct {
		Model   string `json:"model"`
		Answers map[string]struct {
			Choice           string  `json:"choice"`
			Confidence       float64 `json:"confidence"`
			AnswerConfidence float64 `json:"answer_confidence"`
		} `json:"answers"`
		Usage struct {
			Truncated json.RawMessage `json:"truncated"`
		} `json:"usage"`
	}
	if err := json.NewDecoder(io.LimitReader(r, 1<<20)).Decode(&decoded); err != nil {
		return Parsed{}, errors.New("laya response unreadable")
	}
	if usageCut(decoded.Usage.Truncated) {
		return Parsed{}, ErrTruncated
	}
	ans, ok := decoded.Answers["verdict"]
	if !ok || strings.TrimSpace(ans.Choice) == "" {
		return Parsed{}, errors.New("laya verdict choice missing")
	}
	return Parsed{
		Choice:           ans.Choice,
		Confidence:       ans.Confidence,
		AnswerConfidence: ans.AnswerConfidence,
		Model:            decoded.Model,
	}, nil
}

func usageCut(raw json.RawMessage) bool {
	trimmed := bytes.TrimSpace(raw)
	if len(trimmed) == 0 || bytes.Equal(trimmed, []byte("null")) {
		return false
	}
	var flag bool
	if err := json.Unmarshal(trimmed, &flag); err == nil {
		return flag
	}
	var n float64
	if err := json.Unmarshal(trimmed, &n); err == nil {
		return n > 0
	}
	return false
}

func detailText(b []byte) string {
	var wrap struct {
		Detail any `json:"detail"`
	}
	if json.Unmarshal(b, &wrap) == nil && wrap.Detail != nil {
		switch v := wrap.Detail.(type) {
		case string:
			return v
		default:
			raw, err := json.Marshal(v)
			if err == nil {
				return string(raw)
			}
		}
	}
	return string(b)
}

func scrub(msg, secret, body string) string {
	if secret != "" {
		msg = strings.ReplaceAll(msg, secret, "redacted")
		msg = strings.ReplaceAll(msg, "Bearer "+secret, "redacted")
	}
	if len(body) >= 8 {
		msg = strings.ReplaceAll(msg, body, "redacted")
	}
	msg = strings.NewReplacer("<", "", ">", "", "\n", " ", "\r", " ", "|", "/").Replace(msg)
	if len(msg) > 180 {
		msg = msg[:180]
	}
	return msg
}

// bareHTTPURL is true for an http(s) URL with no path, query, fragment, or opaque part.
func bareHTTPURL(u *url.URL) bool {
	if u.Scheme != "http" && u.Scheme != "https" {
		return false
	}
	if u.Path != "" && u.Path != "/" {
		return false
	}
	return u.RawQuery == "" && u.Fragment == "" && u.Opaque == ""
}

func (c *Client) newRequest(ctx context.Context, endpoint, body string, criteria map[string]string) (*http.Request, error) {
	if ctx == nil {
		ctx = context.Background()
	}
	payload := Request{
		State: State{Body: body},
		Questions: map[string]Question{
			"verdict": {
				Type:         "choice",
				Instructions: Instructions,
				Criteria:     criteria,
			},
		},
		Model: modelEnglish,
	}
	raw, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.New("laya response unreadable")
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(raw))
	if err != nil {
		return nil, fmt.Errorf("%w: %s", ErrTransport, scrub(err.Error(), c.apiKey, body))
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}
