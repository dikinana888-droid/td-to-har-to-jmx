package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/tcpdump-to-jmx/internal/converter"
	"github.com/tcpdump-to-jmx/internal/models"
)

var (
	// Input/Output flags
	inputFile  string
	outputDir  string
	outputType string

	// Filter flags
	filterPort int
	filterHost string

	// JMX conversion flags
	enableCorrelation      bool
	enableParameterization bool
	threadCount            string
	rampUpTime             string
	loopCount              string

	// Output format flags
	prettyPrint bool
)

// convertCmd represents the convert command
var convertCmd = &cobra.Command{
	Use:   "convert",
	Short: "Convert PCAP file to HAR and/or JMX format",
	Long: `Convert a PCAP (packet capture) file to HAR (HTTP Archive) and/or JMX (JMeter Test Plan) formats.

The converter extracts HTTP traffic from the PCAP file and generates:
- HAR file: Contains all HTTP requests and responses in a standard format
- JMX file: Creates a JMeter test plan that can replay the captured traffic

Examples:
  # Convert PCAP to both HAR and JMX
  tcpdump-to-jmx convert -i capture.pcap -o ./output

  # Convert PCAP to HAR only
  tcpdump-to-jmx convert -i capture.pcap -o ./output -t har

  # Convert PCAP to JMX with specific settings
  tcpdump-to-jmx convert -i capture.pcap -o ./output -t jmx --threads 10 --ramp-up 5

  # Filter by port and host
  tcpdump-to-jmx convert -i capture.pcap -o ./output --port 8080 --host example.com`,
	RunE: runConvert,
}

func init() {
	rootCmd.AddCommand(convertCmd)

	// Input/Output flags
	convertCmd.Flags().StringVarP(&inputFile, "input", "i", "", "Input PCAP file path (required)")
	convertCmd.Flags().StringVarP(&outputDir, "output", "o", ".", "Output directory for generated files")
	convertCmd.Flags().StringVarP(&outputType, "type", "t", "both", "Output type: har, jmx, or both")

	// Filter flags
	convertCmd.Flags().IntVar(&filterPort, "port", 0, "Filter traffic by port (0 = all ports)")
	convertCmd.Flags().StringVar(&filterHost, "host", "", "Filter traffic by host")

	// JMX conversion flags
	convertCmd.Flags().BoolVar(&enableCorrelation, "correlation", false, "Enable automatic correlation detection in JMX")
	convertCmd.Flags().BoolVar(&enableParameterization, "parameterization", false, "Enable automatic parameterization in JMX")
	convertCmd.Flags().StringVar(&threadCount, "threads", "1", "Number of threads in JMX test plan")
	convertCmd.Flags().StringVar(&rampUpTime, "ramp-up", "1", "Ramp-up time in seconds for JMX test plan")
	convertCmd.Flags().StringVar(&loopCount, "loops", "1", "Number of loops in JMX test plan")

	// Output format flags
	convertCmd.Flags().BoolVar(&prettyPrint, "pretty", true, "Pretty print JSON output")

	// Mark required flags
	convertCmd.MarkFlagRequired("input")
}

func runConvert(cmd *cobra.Command, args []string) error {
	// Validate input file
	if _, err := os.Stat(inputFile); os.IsNotExist(err) {
		return fmt.Errorf("input file does not exist: %s", inputFile)
	}

	// Validate output type
	outputType = strings.ToLower(outputType)
	if outputType != "har" && outputType != "jmx" && outputType != "both" {
		return fmt.Errorf("invalid output type: %s (must be har, jmx, or both)", outputType)
	}

	// Create output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %v", err)
	}

	// Read PCAP file
	logrus.Infof("Reading PCAP file: %s", inputFile)
	pcapData, err := ioutil.ReadFile(inputFile)
	if err != nil {
		return fmt.Errorf("failed to read PCAP file: %v", err)
	}

	// Get base filename without extension
	baseFilename := strings.TrimSuffix(filepath.Base(inputFile), filepath.Ext(inputFile))

	// Convert PCAP to HAR
	logrus.Info("Converting PCAP to HAR format...")
	pcapConverter := converter.NewPcapToHarConverter()
	
	// Apply filters if specified
	if filterPort > 0 {
		pcapConverter.SetPortFilter(filterPort)
		logrus.Debugf("Filtering by port: %d", filterPort)
	}
	if filterHost != "" {
		pcapConverter.SetHostFilter(filterHost)
		logrus.Debugf("Filtering by host: %s", filterHost)
	}

	// Convert to HAR
	har, err := pcapConverter.Convert(pcapData)
	if err != nil {
		return fmt.Errorf("failed to convert PCAP to HAR: %v", err)
	}

	if len(har.Log.Entries) == 0 {
		logrus.Warn("No HTTP traffic found in PCAP file")
		return nil
	}

	logrus.Infof("Found %d HTTP requests in PCAP file", len(har.Log.Entries))

	// Save HAR file if requested
	if outputType == "har" || outputType == "both" {
		harPath := filepath.Join(outputDir, baseFilename+".har")
		if err := saveHAR(har, harPath); err != nil {
			return fmt.Errorf("failed to save HAR file: %v", err)
		}
		logrus.Infof("HAR file saved: %s", harPath)
	}

	// Convert to JMX if requested
	if outputType == "jmx" || outputType == "both" {
		logrus.Info("Converting HAR to JMX format...")
		
		// Create conversion options
		options := models.ConversionOptions{
			EnableCorrelation:      enableCorrelation,
			EnableParameterization: enableParameterization,
			ThreadCount:            threadCount,
			RampUpTime:             rampUpTime,
			LoopCount:              loopCount,
		}

		// Convert HAR to JMX
		jmxConverter := converter.NewHarToJmxConverter(options)
		jmxData, err := jmxConverter.Convert(har)
		if err != nil {
			return fmt.Errorf("failed to convert HAR to JMX: %v", err)
		}

		// Save JMX file
		jmxPath := filepath.Join(outputDir, baseFilename+".jmx")
		if err := saveJMX(jmxData, jmxPath); err != nil {
			return fmt.Errorf("failed to save JMX file: %v", err)
		}
		logrus.Infof("JMX file saved: %s", jmxPath)

		// Print conversion summary
		printConversionSummary(har, options)
	}

	logrus.Info("Conversion completed successfully!")
	return nil
}

func saveHAR(har *models.HAR, filepath string) error {
	var data []byte
	var err error

	if prettyPrint {
		data, err = json.MarshalIndent(har, "", "  ")
	} else {
		data, err = json.Marshal(har)
	}

	if err != nil {
		return err
	}

	return ioutil.WriteFile(filepath, data, 0644)
}

func saveJMX(jmxData []byte, filepath string) error {
	return ioutil.WriteFile(filepath, jmxData, 0644)
}

func printConversionSummary(har *models.HAR, options models.ConversionOptions) {
	fmt.Println("\n=== Conversion Summary ===")
	fmt.Printf("Total Requests: %d\n", len(har.Log.Entries))
	
	// Count unique hosts
	hosts := make(map[string]int)
	for _, entry := range har.Log.Entries {
		host := entry.Request.URL
		if idx := strings.Index(host, "://"); idx != -1 {
			host = host[idx+3:]
			if idx := strings.Index(host, "/"); idx != -1 {
				host = host[:idx]
			}
		}
		hosts[host]++
	}
	
	fmt.Printf("Unique Hosts: %d\n", len(hosts))
	
	// Show top hosts
	if len(hosts) > 0 && verbose {
		fmt.Println("\nTop Hosts:")
		for host, count := range hosts {
			fmt.Printf("  - %s: %d requests\n", host, count)
		}
	}
	
	// Show JMX settings
	fmt.Println("\nJMX Settings:")
	fmt.Printf("  Thread Count: %s\n", options.ThreadCount)
	fmt.Printf("  Ramp-up Time: %s seconds\n", options.RampUpTime)
	fmt.Printf("  Loop Count: %s\n", options.LoopCount)
	fmt.Printf("  Correlation: %v\n", options.EnableCorrelation)
	fmt.Printf("  Parameterization: %v\n", options.EnableParameterization)
}