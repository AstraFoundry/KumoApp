// Package fetch downloads remote profiles and overrides.
package fetch

import (
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/AstraFoundry/KumoApp/internal/domain/profile"
)

// UserAgent identifies Kumo as a Clash-compatible client, so subscription
// servers return Mihomo YAML and their usage headers.
const UserAgent = "clash.meta/Kumo"

// Client fetches over HTTP(S).
type Client struct {
	Timeout time.Duration
}

// New returns a client with a 30 second timeout.
func New() *Client { return &Client{Timeout: 30 * time.Second} }

func (c *Client) http(proxyPort int) *http.Client {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	if proxyPort > 0 {
		proxyURL := &url.URL{Scheme: "http", Host: "127.0.0.1:" + strconv.Itoa(proxyPort)}
		transport.Proxy = http.ProxyURL(proxyURL)
	}
	return &http.Client{Timeout: c.Timeout, Transport: transport}
}

// Get downloads a URL's body.
func (c *Client) Get(rawURL string, proxyPort int) ([]byte, http.Header, error) {
	req, err := http.NewRequest(http.MethodGet, rawURL, nil)
	if err != nil {
		return nil, nil, err
	}
	req.Header.Set("User-Agent", UserAgent)
	resp, err := c.http(proxyPort).Do(req)
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return nil, nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		return nil, nil, fmt.Errorf("%s answered %s", hostOf(rawURL), resp.Status)
	}
	return body, resp.Header, nil
}

// Fetch downloads a profile and reads the headers subscriptions send.
func (c *Client) Fetch(rawURL string, proxyPort int) (profile.Document, error) {
	body, header, err := c.Get(rawURL, proxyPort)
	if err != nil {
		return profile.Document{}, err
	}
	doc := profile.Document{YAML: string(body), Name: filename(header.Get("Content-Disposition"))}
	if doc.Name == "" {
		doc.Name = hostOf(rawURL)
	}
	doc.HomeURL = headerSuffix(header, "profile-web-page-url")
	if minutes, err := strconv.ParseFloat(headerSuffix(header, "profile-update-interval"), 64); err == nil && minutes > 0 {
		doc.UpdateIntervalSeconds = int(minutes * 3600)
	}
	if usage, ok := profile.ParseUsage(headerSuffix(header, "subscription-userinfo")); ok {
		doc.Usage = &usage
	}
	return doc, nil
}

// headerSuffix finds a header by the end of its name, as providers prefix
// them inconsistently.
func headerSuffix(header http.Header, suffix string) string {
	for key, values := range header {
		if strings.HasSuffix(strings.ToLower(key), suffix) && len(values) > 0 {
			return values[0]
		}
	}
	return ""
}

func filename(disposition string) string {
	if disposition == "" {
		return ""
	}
	_, params, err := mime.ParseMediaType(disposition)
	if err == nil && params["filename"] != "" {
		return strings.TrimSuffix(strings.TrimSuffix(params["filename"], ".yaml"), ".yml")
	}
	if i := strings.Index(strings.ToLower(disposition), "filename*=utf-8''"); i >= 0 {
		name, _ := url.PathUnescape(disposition[i+len("filename*=utf-8''"):])
		return name
	}
	return ""
}

func hostOf(rawURL string) string {
	if u, err := url.Parse(rawURL); err == nil && u.Host != "" {
		return u.Host
	}
	return "Remote profile"
}
