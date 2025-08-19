package converter

import "encoding/xml"

// JMeterTestPlan represents the root JMeter test plan
type JMeterTestPlan struct {
	XMLName    xml.Name  `xml:"jmeterTestPlan"`
	Version    string    `xml:"version,attr"`
	Properties string    `xml:"properties,attr"`
	JMeter     string    `xml:"jmeter,attr"`
	HashTree   *HashTree `xml:"hashTree"`
}

// HashTree represents a JMeter hash tree
type HashTree struct {
	TestPlan    TestPlan          `xml:"TestPlan,omitempty"`
	HashTree    []HashTreeElement `xml:"hashTree,omitempty"`
}

// HashTreeElement represents an element in the hash tree
type HashTreeElement struct {
	ThreadGroup     *ThreadGroup     `xml:"ThreadGroup,omitempty"`
	HTTPSampler     *HTTPSampler     `xml:"HTTPSamplerProxy,omitempty"`
	HeaderManager   *HeaderManager   `xml:"HeaderManager,omitempty"`
	CookieManager   *CookieManager   `xml:"CookieManager,omitempty"`
	ResultCollector *ResultCollector `xml:"ResultCollector,omitempty"`
	RegexExtractor  *RegexExtractor  `xml:"RegexExtractor,omitempty"`
	HashTree        *HashTree        `xml:"hashTree,omitempty"`
}

// TestPlan represents a JMeter test plan element
type TestPlan struct {
	Guiclass    string      `xml:"guiclass,attr"`
	Testclass   string      `xml:"testclass,attr"`
	Testname    string      `xml:"testname,attr"`
	Enabled     string      `xml:"enabled,attr"`
	StringProps []StringProp `xml:"stringProp,omitempty"`
	BoolProps   []BoolProp   `xml:"boolProp,omitempty"`
	ElementProp ElementProp  `xml:"elementProp,omitempty"`
}

// ThreadGroup represents a JMeter thread group
type ThreadGroup struct {
	Guiclass    string       `xml:"guiclass,attr"`
	Testclass   string       `xml:"testclass,attr"`
	Testname    string       `xml:"testname,attr"`
	Enabled     string       `xml:"enabled,attr"`
	StringProps []StringProp `xml:"stringProp,omitempty"`
	BoolProps   []BoolProp   `xml:"boolProp,omitempty"`
	ElementProp ElementProp  `xml:"elementProp,omitempty"`
}

// HTTPSampler represents a JMeter HTTP sampler
type HTTPSampler struct {
	Guiclass    string      `xml:"guiclass,attr"`
	Testclass   string      `xml:"testclass,attr"`
	Testname    string      `xml:"testname,attr"`
	Enabled     string      `xml:"enabled,attr"`
	StringProps []StringProp `xml:"stringProp,omitempty"`
	BoolProps   []BoolProp   `xml:"boolProp,omitempty"`
	ElementProp ElementProp  `xml:"elementProp,omitempty"`
}

// HeaderManager represents a JMeter HTTP header manager
type HeaderManager struct {
	Guiclass       string         `xml:"guiclass,attr"`
	Testclass      string         `xml:"testclass,attr"`
	Testname       string         `xml:"testname,attr"`
	Enabled        string         `xml:"enabled,attr"`
	CollectionProp CollectionProp `xml:"collectionProp,omitempty"`
}

// CookieManager represents a JMeter HTTP cookie manager
type CookieManager struct {
	Guiclass       string         `xml:"guiclass,attr"`
	Testclass      string         `xml:"testclass,attr"`
	Testname       string         `xml:"testname,attr"`
	Enabled        string         `xml:"enabled,attr"`
	BoolProps      []BoolProp     `xml:"boolProp,omitempty"`
	CollectionProp CollectionProp `xml:"collectionProp,omitempty"`
}

// ResultCollector represents a JMeter result collector (listener)
type ResultCollector struct {
	Guiclass    string       `xml:"guiclass,attr"`
	Testclass   string       `xml:"testclass,attr"`
	Testname    string       `xml:"testname,attr"`
	Enabled     string       `xml:"enabled,attr"`
	BoolProps   []BoolProp   `xml:"boolProp,omitempty"`
	StringProps []StringProp `xml:"stringProp,omitempty"`
}

// RegexExtractor represents a JMeter regular expression extractor
type RegexExtractor struct {
	Guiclass    string       `xml:"guiclass,attr"`
	Testclass   string       `xml:"testclass,attr"`
	Testname    string       `xml:"testname,attr"`
	Enabled     string       `xml:"enabled,attr"`
	StringProps []StringProp `xml:"stringProp,omitempty"`
}

// ElementProp represents a JMeter element property
type ElementProp struct {
	Name           string         `xml:"name,attr"`
	ElementType    string         `xml:"elementType,attr,omitempty"`
	Guiclass       string         `xml:"guiclass,attr,omitempty"`
	Testclass      string         `xml:"testclass,attr,omitempty"`
	Enabled        string         `xml:"enabled,attr,omitempty"`
	StringProps    []StringProp   `xml:"stringProp,omitempty"`
	BoolProps      []BoolProp     `xml:"boolProp,omitempty"`
	CollectionProp CollectionProp `xml:"collectionProp,omitempty"`
}

// CollectionProp represents a JMeter collection property
type CollectionProp struct {
	Name         string        `xml:"name,attr"`
	ElementProps []ElementProp `xml:"elementProp,omitempty"`
}

// StringProp represents a JMeter string property
type StringProp struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}

// BoolProp represents a JMeter boolean property
type BoolProp struct {
	Name  string `xml:"name,attr"`
	Value string `xml:",chardata"`
}