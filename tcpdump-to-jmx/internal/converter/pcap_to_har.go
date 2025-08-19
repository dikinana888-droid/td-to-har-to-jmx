package converter

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/google/gopacket"
	"github.com/google/gopacket/layers"
	"github.com/google/gopacket/pcap"
	"github.com/google/gopacket/tcpassembly"
	"github.com/google/gopacket/tcpassembly/tcpreader"
	"github.com/sirupsen/logrus"
	"github.com/tcpdump-to-jmx/internal/models"
)

// PcapToHarConverter converts PCAP files to HAR format
type PcapToHarConverter struct {
	filterPort int
	filterHost string
	entries    []models.HAREntry
	startTime  time.Time
}

// NewPcapToHarConverter creates a new PCAP to HAR converter
func NewPcapToHarConverter() *PcapToHarConverter {
	return &PcapToHarConverter{
		entries: make([]models.HAREntry, 0),
	}
}

// SetPortFilter sets the port filter
func (c *PcapToHarConverter) SetPortFilter(port int) {
	c.filterPort = port
}

// SetHostFilter sets the host filter
func (c *PcapToHarConverter) SetHostFilter(host string) {
	c.filterHost = host
}

// Convert converts PCAP data to HAR format
func (c *PcapToHarConverter) Convert(pcapData []byte) (*models.HAR, error) {
	// Write PCAP data to temporary file
	tmpFile, err := os.CreateTemp("", "pcap-*.pcap")
	if err != nil {
		return nil, fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	if _, err := tmpFile.Write(pcapData); err != nil {
		return nil, fmt.Errorf("failed to write PCAP data: %w", err)
	}

	// Create a packet source from the PCAP file
	handle, err := pcap.OpenOffline(tmpFile.Name())
	if err != nil {
		return nil, fmt.Errorf("failed to open PCAP: %w", err)
	}
	defer handle.Close()

	// Set up BPF filter if needed
	if c.filterPort > 0 || c.filterHost != "" {
		filter := c.buildBPFFilter()
		if filter != "" {
			if err := handle.SetBPFFilter(filter); err != nil {
				logrus.Warnf("Failed to set BPF filter: %v", err)
			}
		}
	}

	// Create packet source
	packetSource := gopacket.NewPacketSource(handle, handle.LinkType())
	
	// Set up TCP assembly
	streamFactory := &httpStreamFactory{converter: c}
	streamPool := tcpassembly.NewStreamPool(streamFactory)
	assembler := tcpassembly.NewAssembler(streamPool)

	// Process packets
	for packet := range packetSource.Packets() {
		if packet.NetworkLayer() == nil || packet.TransportLayer() == nil {
			continue
		}

		tcp, ok := packet.TransportLayer().(*layers.TCP)
		if !ok {
			continue
		}

		// Assemble TCP streams
		assembler.AssembleWithTimestamp(
			packet.NetworkLayer().NetworkFlow(),
			tcp,
			packet.Metadata().Timestamp,
		)
	}

	// Flush remaining data
	assembler.FlushAll()

	// Create HAR structure
	har := &models.HAR{
		Log: models.HARLog{
			Version: "1.2",
			Creator: models.HARCreator{
				Name:    "tcpdump-to-jmx",
				Version: "1.0.0",
			},
			Entries: c.entries,
		},
	}

	return har, nil
}

func (c *PcapToHarConverter) buildBPFFilter() string {
	var filters []string
	
	if c.filterPort > 0 {
		filters = append(filters, fmt.Sprintf("port %d", c.filterPort))
	}
	
	if c.filterHost != "" {
		filters = append(filters, fmt.Sprintf("host %s", c.filterHost))
	}
	
	return strings.Join(filters, " and ")
}

// httpStreamFactory implements tcpassembly.StreamFactory
type httpStreamFactory struct {
	converter *PcapToHarConverter
}

func (f *httpStreamFactory) New(net, transport gopacket.Flow) tcpassembly.Stream {
	stream := &httpStream{
		net:       net,
		transport: transport,
		converter: f.converter,
		reader:    tcpreader.NewReaderStream(),
	}
	go stream.run()
	return &stream.reader
}

// httpStream handles HTTP stream assembly
type httpStream struct {
	net       gopacket.Flow
	transport gopacket.Flow
	converter *PcapToHarConverter
	reader    tcpreader.ReaderStream
}

func (s *httpStream) run() {
	defer s.reader.Close()
	
	buf := new(bytes.Buffer)
	io.Copy(buf, &s.reader)
	
	data := buf.Bytes()
	if len(data) == 0 {
		return
	}

	// Try to parse as HTTP request
	if req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(data))); err == nil {
		s.handleRequest(req, data)
		return
	}

	// Try to parse as HTTP response
	if strings.HasPrefix(string(data), "HTTP/") {
		s.handleResponse(data)
	}
}

func (s *httpStream) handleRequest(req *http.Request, rawData []byte) {
	// Create HAR entry
	entry := models.HAREntry{
		StartedDateTime: time.Now(),
		Request: models.HARRequest{
			Method:      req.Method,
			URL:         s.buildURL(req),
			HTTPVersion: req.Proto,
			Headers:     s.convertHeaders(req.Header),
			QueryString: s.parseQueryString(req.URL.Query()),
			HeadersSize: len(rawData),
			BodySize:    int(req.ContentLength),
		},
		Response: models.HARResponse{
			Status:      0,
			StatusText:  "",
			HTTPVersion: "",
			Headers:     []models.HARHeader{},
			Content:     models.HARContent{},
			RedirectURL: "",
			HeadersSize: -1,
			BodySize:    -1,
		},
		Cache: models.HARCache{},
		Timings: models.HARTimings{
			Send:    1,
			Wait:    1,
			Receive: 1,
		},
	}

	// Handle POST data
	if req.Method == "POST" || req.Method == "PUT" || req.Method == "PATCH" {
		body, _ := io.ReadAll(req.Body)
		req.Body = io.NopCloser(bytes.NewReader(body))
		
		entry.Request.PostData = &models.HARPostData{
			MimeType: req.Header.Get("Content-Type"),
			Text:     string(body),
		}
	}

	// Handle cookies
	for _, cookie := range req.Cookies() {
		entry.Request.Cookies = append(entry.Request.Cookies, models.HARCookie{
			Name:  cookie.Name,
			Value: cookie.Value,
			Path:  cookie.Path,
			Domain: cookie.Domain,
		})
	}

	s.converter.entries = append(s.converter.entries, entry)
}

func (s *httpStream) handleResponse(data []byte) {
	// Parse response
	resp, err := http.ReadResponse(bufio.NewReader(bytes.NewReader(data)), nil)
	if err != nil {
		return
	}

	// Find corresponding request entry and update it
	if len(s.converter.entries) > 0 {
		lastEntry := &s.converter.entries[len(s.converter.entries)-1]
		
		lastEntry.Response = models.HARResponse{
			Status:      resp.StatusCode,
			StatusText:  resp.Status,
			HTTPVersion: resp.Proto,
			Headers:     s.convertHeaders(resp.Header),
			RedirectURL: resp.Header.Get("Location"),
			HeadersSize: len(data),
			BodySize:    int(resp.ContentLength),
		}

		// Read response body
		body, _ := io.ReadAll(resp.Body)
		resp.Body = io.NopCloser(bytes.NewReader(body))
		
		// Handle compressed content
		contentEncoding := resp.Header.Get("Content-Encoding")
		if contentEncoding == "gzip" {
			if reader, err := gzip.NewReader(bytes.NewReader(body)); err == nil {
				decompressed, _ := io.ReadAll(reader)
				lastEntry.Response.Content = models.HARContent{
					Size:        len(decompressed),
					Compression: len(body) - len(decompressed),
					MimeType:    resp.Header.Get("Content-Type"),
					Text:        string(decompressed),
					Encoding:    contentEncoding,
				}
			}
		} else {
			lastEntry.Response.Content = models.HARContent{
				Size:     len(body),
				MimeType: resp.Header.Get("Content-Type"),
				Text:     string(body),
			}
		}

		// Calculate timing
		lastEntry.Time = float64(time.Since(lastEntry.StartedDateTime).Milliseconds())
	}
}

func (s *httpStream) buildURL(req *http.Request) string {
	scheme := "http"
	if req.TLS != nil {
		scheme = "https"
	}
	
	host := req.Host
	if host == "" {
		host = s.transport.Dst().String()
	}
	
	return fmt.Sprintf("%s://%s%s", scheme, host, req.URL.String())
}

func (s *httpStream) convertHeaders(headers http.Header) []models.HARHeader {
	result := make([]models.HARHeader, 0, len(headers))
	for name, values := range headers {
		for _, value := range values {
			result = append(result, models.HARHeader{
				Name:  name,
				Value: value,
			})
		}
	}
	return result
}

func (s *httpStream) parseQueryString(values url.Values) []models.HARParam {
	result := make([]models.HARParam, 0, len(values))
	for name, vals := range values {
		for _, value := range vals {
			result = append(result, models.HARParam{
				Name:  name,
				Value: value,
			})
		}
	}
	return result
}