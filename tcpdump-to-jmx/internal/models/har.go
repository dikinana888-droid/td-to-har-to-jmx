package models

import (
	"time"
)

// HAR represents the HTTP Archive format
type HAR struct {
	Log HARLog `json:"log"`
}

// HARLog represents the log section of HAR
type HARLog struct {
	Version string       `json:"version"`
	Creator HARCreator   `json:"creator"`
	Browser HARBrowser   `json:"browser,omitempty"`
	Pages   []HARPage    `json:"pages,omitempty"`
	Entries []HAREntry   `json:"entries"`
	Comment string       `json:"comment,omitempty"`
}

// HARCreator represents the creator of the HAR file
type HARCreator struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Comment string `json:"comment,omitempty"`
}

// HARBrowser represents browser information
type HARBrowser struct {
	Name    string `json:"name"`
	Version string `json:"version"`
	Comment string `json:"comment,omitempty"`
}

// HARPage represents a page in the HAR file
type HARPage struct {
	StartedDateTime time.Time       `json:"startedDateTime"`
	ID              string          `json:"id"`
	Title           string          `json:"title"`
	PageTimings     HARPageTimings  `json:"pageTimings"`
	Comment         string          `json:"comment,omitempty"`
}

// HARPageTimings represents page load timings
type HARPageTimings struct {
	OnContentLoad int `json:"onContentLoad,omitempty"`
	OnLoad        int `json:"onLoad,omitempty"`
	Comment       string `json:"comment,omitempty"`
}

// HAREntry represents a single HTTP transaction
type HAREntry struct {
	PageRef         string       `json:"pageref,omitempty"`
	StartedDateTime time.Time    `json:"startedDateTime"`
	Time            float64      `json:"time"`
	Request         HARRequest   `json:"request"`
	Response        HARResponse  `json:"response"`
	Cache           HARCache     `json:"cache"`
	Timings         HARTimings   `json:"timings"`
	ServerIPAddress string       `json:"serverIPAddress,omitempty"`
	Connection      string       `json:"connection,omitempty"`
	Comment         string       `json:"comment,omitempty"`
}

// HARRequest represents an HTTP request
type HARRequest struct {
	Method      string         `json:"method"`
	URL         string         `json:"url"`
	HTTPVersion string         `json:"httpVersion"`
	Cookies     []HARCookie    `json:"cookies"`
	Headers     []HARHeader    `json:"headers"`
	QueryString []HARParam     `json:"queryString"`
	PostData    *HARPostData   `json:"postData,omitempty"`
	HeadersSize int            `json:"headersSize"`
	BodySize    int            `json:"bodySize"`
	Comment     string         `json:"comment,omitempty"`
}

// HARResponse represents an HTTP response
type HARResponse struct {
	Status      int         `json:"status"`
	StatusText  string      `json:"statusText"`
	HTTPVersion string      `json:"httpVersion"`
	Cookies     []HARCookie `json:"cookies"`
	Headers     []HARHeader `json:"headers"`
	Content     HARContent  `json:"content"`
	RedirectURL string      `json:"redirectURL"`
	HeadersSize int         `json:"headersSize"`
	BodySize    int         `json:"bodySize"`
	Comment     string      `json:"comment,omitempty"`
}

// HARCookie represents an HTTP cookie
type HARCookie struct {
	Name     string    `json:"name"`
	Value    string    `json:"value"`
	Path     string    `json:"path,omitempty"`
	Domain   string    `json:"domain,omitempty"`
	Expires  time.Time `json:"expires,omitempty"`
	HTTPOnly bool      `json:"httpOnly,omitempty"`
	Secure   bool      `json:"secure,omitempty"`
	Comment  string    `json:"comment,omitempty"`
}

// HARHeader represents an HTTP header
type HARHeader struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Comment string `json:"comment,omitempty"`
}

// HARParam represents a query parameter
type HARParam struct {
	Name    string `json:"name"`
	Value   string `json:"value"`
	Comment string `json:"comment,omitempty"`
}

// HARPostData represents POST data
type HARPostData struct {
	MimeType string      `json:"mimeType"`
	Params   []HARParam  `json:"params,omitempty"`
	Text     string      `json:"text,omitempty"`
	Comment  string      `json:"comment,omitempty"`
}

// HARContent represents response content
type HARContent struct {
	Size        int    `json:"size"`
	Compression int    `json:"compression,omitempty"`
	MimeType    string `json:"mimeType"`
	Text        string `json:"text,omitempty"`
	Encoding    string `json:"encoding,omitempty"`
	Comment     string `json:"comment,omitempty"`
}

// HARCache represents cache information
type HARCache struct {
	BeforeRequest *HARCacheEntry `json:"beforeRequest,omitempty"`
	AfterRequest  *HARCacheEntry `json:"afterRequest,omitempty"`
	Comment       string         `json:"comment,omitempty"`
}

// HARCacheEntry represents a cache entry
type HARCacheEntry struct {
	Expires    time.Time `json:"expires,omitempty"`
	LastAccess time.Time `json:"lastAccess,omitempty"`
	ETag       string    `json:"eTag,omitempty"`
	HitCount   int       `json:"hitCount,omitempty"`
	Comment    string    `json:"comment,omitempty"`
}

// HARTimings represents request/response timings
type HARTimings struct {
	Blocked int     `json:"blocked,omitempty"`
	DNS     int     `json:"dns,omitempty"`
	Connect int     `json:"connect,omitempty"`
	Send    int     `json:"send"`
	Wait    int     `json:"wait"`
	Receive int     `json:"receive"`
	SSL     int     `json:"ssl,omitempty"`
	Comment string  `json:"comment,omitempty"`
}