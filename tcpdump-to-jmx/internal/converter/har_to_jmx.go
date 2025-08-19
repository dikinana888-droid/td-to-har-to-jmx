package converter

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/tcpdump-to-jmx/internal/models"
)

// HarToJmxConverter converts HAR files to JMX format
type HarToJmxConverter struct {
	enableCorrelation      bool
	enableParameterization bool
	threadCount            string
	rampUpTime             string
	loopCount              string
	correlations           map[string]*Correlation
	parameters             map[string]*Parameter
}

// Correlation represents a correlated value
type Correlation struct {
	Name         string
	Pattern      string
	ExtractFrom  string // response body, header, or cookie
	DefaultValue string
}

// Parameter represents a parameterized value
type Parameter struct {
	Name         string
	Values       []string
	CurrentIndex int
}

// NewHarToJmxConverter creates a new HAR to JMX converter
func NewHarToJmxConverter(options models.ConversionOptions) *HarToJmxConverter {
	return &HarToJmxConverter{
		enableCorrelation:      options.EnableCorrelation,
		enableParameterization: options.EnableParameterization,
		threadCount:            options.ThreadCount,
		rampUpTime:             options.RampUpTime,
		loopCount:              options.LoopCount,
		correlations:           make(map[string]*Correlation),
		parameters:             make(map[string]*Parameter),
	}
}

// Convert converts HAR to JMX format
func (c *HarToJmxConverter) Convert(har *models.HAR) ([]byte, error) {
	// Create JMeter test plan
	testPlan := c.createTestPlan(har)
	
	// Marshal to XML
	output, err := xml.MarshalIndent(testPlan, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("failed to marshal JMX: %w", err)
	}
	
	// Add XML header
	result := []byte(xml.Header + string(output))
	
	return result, nil
}

func (c *HarToJmxConverter) createTestPlan(har *models.HAR) *JMeterTestPlan {
	testPlan := &JMeterTestPlan{
		Version:    "1.2",
		Properties: "5.0",
		JMeter:     "5.5",
	}
	
	// Create hash tree
	hashTree := &HashTree{
		TestPlan: TestPlan{
			Guiclass:    "TestPlanGui",
			Testclass:   "TestPlan",
			Testname:    "Converted Test Plan",
			Enabled:     "true",
			ElementProp: c.createTestPlanArguments(),
		},
	}
	
	// Add thread group
	threadGroup := c.createThreadGroup()
	hashTree.HashTree = append(hashTree.HashTree, HashTreeElement{
		ThreadGroup: &threadGroup,
		HashTree:    c.createHTTPSamplers(har),
	})
	
	// Add listeners
	hashTree.HashTree = append(hashTree.HashTree, c.createListeners()...)
	
	testPlan.HashTree = hashTree
	
	return testPlan
}

func (c *HarToJmxConverter) createTestPlanArguments() ElementProp {
	return ElementProp{
		Name:        "TestPlan.user_defined_variables",
		ElementType: "Arguments",
		Guiclass:    "ArgumentsPanel",
		Testclass:   "Arguments",
		Enabled:     "true",
		CollectionProp: CollectionProp{
			Name: "Arguments.arguments",
		},
	}
}

func (c *HarToJmxConverter) createThreadGroup() ThreadGroup {
	return ThreadGroup{
		Guiclass:  "ThreadGroupGui",
		Testclass: "ThreadGroup",
		Testname:  "Thread Group",
		Enabled:   "true",
		StringProps: []StringProp{
			{Name: "ThreadGroup.on_sample_error", Value: "continue"},
			{Name: "ThreadGroup.num_threads", Value: c.threadCount},
			{Name: "ThreadGroup.ramp_time", Value: c.rampUpTime},
			{Name: "ThreadGroup.duration", Value: ""},
			{Name: "ThreadGroup.delay", Value: ""},
		},
		BoolProps: []BoolProp{
			{Name: "ThreadGroup.scheduler", Value: "false"},
			{Name: "ThreadGroup.same_user_on_next_iteration", Value: "true"},
		},
		ElementProp: ElementProp{
			Name:        "ThreadGroup.main_controller",
			ElementType: "LoopController",
			Guiclass:    "LoopControlPanel",
			Testclass:   "LoopController",
			Enabled:     "true",
			StringProps: []StringProp{
				{Name: "LoopController.loops", Value: c.loopCount},
			},
			BoolProps: []BoolProp{
				{Name: "LoopController.continue_forever", Value: "false"},
			},
		},
	}
}

func (c *HarToJmxConverter) createHTTPSamplers(har *models.HAR) *HashTree {
	hashTree := &HashTree{}
	
	// Add HTTP Cookie Manager
	hashTree.HashTree = append(hashTree.HashTree, HashTreeElement{
		CookieManager: &CookieManager{
			Guiclass:  "CookiePanel",
			Testclass: "CookieManager",
			Testname:  "HTTP Cookie Manager",
			Enabled:   "true",
			BoolProps: []BoolProp{
				{Name: "CookieManager.clearEachIteration", Value: "false"},
				{Name: "CookieManager.controlledByThreadGroup", Value: "false"},
			},
		},
	})
	
	// Add HTTP Header Manager
	hashTree.HashTree = append(hashTree.HashTree, HashTreeElement{
		HeaderManager: &HeaderManager{
			Guiclass:  "HeaderPanel",
			Testclass: "HeaderManager",
			Testname:  "HTTP Header Manager",
			Enabled:   "true",
			CollectionProp: c.createCommonHeaders(har),
		},
	})
	
	// Process each HAR entry
	for i, entry := range har.Log.Entries {
		sampler := c.createHTTPSampler(entry, i)
		samplerElement := HashTreeElement{
			HTTPSampler: &sampler,
		}
		
		// Add extractors if correlation is enabled
		if c.enableCorrelation {
			extractors := c.detectAndCreateExtractors(entry)
			if len(extractors) > 0 {
				samplerElement.HashTree = &HashTree{
					HashTree: extractors,
				}
			}
		}
		
		hashTree.HashTree = append(hashTree.HashTree, samplerElement)
	}
	
	return hashTree
}

func (c *HarToJmxConverter) createHTTPSampler(entry models.HAREntry, index int) HTTPSampler {
	parsedURL, _ := url.Parse(entry.Request.URL)
	
	// Extract protocol, domain, port, and path
	protocol := parsedURL.Scheme
	domain := parsedURL.Hostname()
	port := parsedURL.Port()
	if port == "" {
		if protocol == "https" {
			port = "443"
		} else {
			port = "80"
		}
	}
	path := parsedURL.Path
	if path == "" {
		path = "/"
	}
	
	sampler := HTTPSampler{
		Guiclass:  "HttpTestSampleGui",
		Testclass: "HTTPSamplerProxy",
		Testname:  fmt.Sprintf("%s %s", entry.Request.Method, path),
		Enabled:   "true",
		StringProps: []StringProp{
			{Name: "HTTPSampler.domain", Value: domain},
			{Name: "HTTPSampler.port", Value: port},
			{Name: "HTTPSampler.protocol", Value: protocol},
			{Name: "HTTPSampler.path", Value: c.parameterizePath(path)},
			{Name: "HTTPSampler.method", Value: entry.Request.Method},
			{Name: "HTTPSampler.contentEncoding", Value: "UTF-8"},
		},
		BoolProps: []BoolProp{
			{Name: "HTTPSampler.follow_redirects", Value: "true"},
			{Name: "HTTPSampler.auto_redirects", Value: "false"},
			{Name: "HTTPSampler.use_keepalive", Value: "true"},
			{Name: "HTTPSampler.DO_MULTIPART_POST", Value: "false"},
		},
	}
	
	// Add query parameters
	if len(entry.Request.QueryString) > 0 {
		sampler.ElementProp = c.createArguments(entry.Request.QueryString)
	}
	
	// Add POST data if present
	if entry.Request.PostData != nil {
		sampler.BoolProps = append(sampler.BoolProps, BoolProp{
			Name:  "HTTPSampler.postBodyRaw",
			Value: "true",
		})
		sampler.ElementProp = ElementProp{
			Name:        "HTTPSampler.Arguments",
			ElementType: "Arguments",
			CollectionProp: CollectionProp{
				Name: "Arguments.arguments",
				ElementProps: []ElementProp{
					{
						Name:        "",
						ElementType: "HTTPArgument",
						StringProps: []StringProp{
							{Name: "Argument.value", Value: c.parameterizeBody(entry.Request.PostData.Text)},
						},
					},
				},
			},
		}
	}
	
	return sampler
}

func (c *HarToJmxConverter) createCommonHeaders(har *models.HAR) CollectionProp {
	headerMap := make(map[string]string)
	
	// Collect common headers from all requests
	for _, entry := range har.Log.Entries {
		for _, header := range entry.Request.Headers {
			// Skip host and connection headers
			if strings.ToLower(header.Name) == "host" || strings.ToLower(header.Name) == "connection" {
				continue
			}
			headerMap[header.Name] = header.Value
		}
	}
	
	// Create collection prop for headers
	collectionProp := CollectionProp{
		Name: "HeaderManager.headers",
	}
	
	for name, value := range headerMap {
		collectionProp.ElementProps = append(collectionProp.ElementProps, ElementProp{
			Name:        "",
			ElementType: "Header",
			StringProps: []StringProp{
				{Name: "Header.name", Value: name},
				{Name: "Header.value", Value: value},
			},
		})
	}
	
	return collectionProp
}

func (c *HarToJmxConverter) detectAndCreateExtractors(entry models.HAREntry) []HashTreeElement {
	var extractors []HashTreeElement
	
	// Detect common correlation patterns in response
	if entry.Response.Content.Text != "" {
		// Session IDs
		if matches := regexp.MustCompile(`(session[_-]?id|jsessionid|phpsessid)["\s:=]+([a-zA-Z0-9\-_]+)`).FindStringSubmatch(entry.Response.Content.Text); len(matches) > 0 {
			extractors = append(extractors, c.createRegexExtractor("sessionId", matches[0], "$1$"))
		}
		
		// CSRF tokens
		if matches := regexp.MustCompile(`(csrf[_-]?token|authenticity[_-]?token|xsrf[_-]?token)["\s:=]+([a-zA-Z0-9\-_]+)`).FindStringSubmatch(entry.Response.Content.Text); len(matches) > 0 {
			extractors = append(extractors, c.createRegexExtractor("csrfToken", matches[0], "$1$"))
		}
		
		// View states (ASP.NET)
		if matches := regexp.MustCompile(`__VIEWSTATE["\s:=]+([a-zA-Z0-9+/=]+)`).FindStringSubmatch(entry.Response.Content.Text); len(matches) > 0 {
			extractors = append(extractors, c.createRegexExtractor("viewState", "__VIEWSTATE[\"\\s:=]+([a-zA-Z0-9+/=]+)", "$1$"))
		}
		
		// JWT tokens
		if matches := regexp.MustCompile(`(token|jwt|bearer)["\s:=]+([a-zA-Z0-9\-_.]+)`).FindStringSubmatch(entry.Response.Content.Text); len(matches) > 0 {
			extractors = append(extractors, c.createRegexExtractor("authToken", matches[0], "$1$"))
		}
	}
	
	return extractors
}

func (c *HarToJmxConverter) createRegexExtractor(varName, regex, template string) HashTreeElement {
	return HashTreeElement{
		RegexExtractor: &RegexExtractor{
			Guiclass:  "RegexExtractorGui",
			Testclass: "RegexExtractor",
			Testname:  fmt.Sprintf("Extract %s", varName),
			Enabled:   "true",
			StringProps: []StringProp{
				{Name: "RegexExtractor.useHeaders", Value: "false"},
				{Name: "RegexExtractor.refname", Value: varName},
				{Name: "RegexExtractor.regex", Value: regex},
				{Name: "RegexExtractor.template", Value: template},
				{Name: "RegexExtractor.default", Value: "NOT_FOUND"},
				{Name: "RegexExtractor.match_number", Value: "1"},
			},
		},
	}
}

func (c *HarToJmxConverter) parameterizePath(path string) string {
	if !c.enableParameterization {
		return path
	}
	
	// Replace numeric IDs with variables
	path = regexp.MustCompile(`/(\d+)`).ReplaceAllString(path, "/${id}")
	
	// Replace UUIDs with variables
	path = regexp.MustCompile(`/([a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12})`).ReplaceAllString(path, "/${uuid}")
	
	return path
}

func (c *HarToJmxConverter) parameterizeBody(body string) string {
	if !c.enableParameterization || body == "" {
		return body
	}
	
	// Try to parse as JSON and parameterize
	var jsonData map[string]interface{}
	if err := json.Unmarshal([]byte(body), &jsonData); err == nil {
		// Parameterize common fields
		for key := range jsonData {
			switch key {
			case "username", "email", "user":
				jsonData[key] = "${username}"
			case "password", "pass", "pwd":
				jsonData[key] = "${password}"
			case "token", "csrf_token", "authenticity_token":
				jsonData[key] = "${" + key + "}"
			default:
				// Keep original value
			}
		}
		
		parameterized, _ := json.Marshal(jsonData)
		return string(parameterized)
	}
	
	// For form data, parameterize common fields
	body = regexp.MustCompile(`(username|email|user)=([^&]+)`).ReplaceAllString(body, "$1=${username}")
	body = regexp.MustCompile(`(password|pass|pwd)=([^&]+)`).ReplaceAllString(body, "$1=${password}")
	
	return body
}

func (c *HarToJmxConverter) createArguments(params []models.HARParam) ElementProp {
	elementProp := ElementProp{
		Name:        "HTTPSampler.Arguments",
		ElementType: "Arguments",
		Guiclass:    "HTTPArgumentsPanel",
		Testclass:   "Arguments",
		Enabled:     "true",
		CollectionProp: CollectionProp{
			Name: "Arguments.arguments",
		},
	}
	
	for _, param := range params {
		paramValue := param.Value
		if c.enableParameterization {
			// Parameterize common parameter values
			if param.Name == "page" || param.Name == "offset" || param.Name == "limit" {
				paramValue = "${" + param.Name + "}"
			}
		}
		
		elementProp.CollectionProp.ElementProps = append(elementProp.CollectionProp.ElementProps, ElementProp{
			Name:        "",
			ElementType: "HTTPArgument",
			StringProps: []StringProp{
				{Name: "Argument.name", Value: param.Name},
				{Name: "Argument.value", Value: paramValue},
				{Name: "Argument.metadata", Value: "="},
			},
			BoolProps: []BoolProp{
				{Name: "HTTPArgument.always_encode", Value: "false"},
				{Name: "HTTPArgument.use_equals", Value: "true"},
			},
		})
	}
	
	return elementProp
}

func (c *HarToJmxConverter) createListeners() []HashTreeElement {
	return []HashTreeElement{
		{
			ResultCollector: &ResultCollector{
				Guiclass:  "ViewResultsFullVisualizer",
				Testclass: "ResultCollector",
				Testname:  "View Results Tree",
				Enabled:   "true",
				BoolProps: []BoolProp{
					{Name: "ResultCollector.error_logging", Value: "false"},
				},
				StringProps: []StringProp{
					{Name: "filename", Value: ""},
				},
			},
		},
		{
			ResultCollector: &ResultCollector{
				Guiclass:  "SummaryReport",
				Testclass: "ResultCollector",
				Testname:  "Summary Report",
				Enabled:   "true",
				BoolProps: []BoolProp{
					{Name: "ResultCollector.error_logging", Value: "false"},
				},
				StringProps: []StringProp{
					{Name: "filename", Value: ""},
				},
			},
		},
	}
}