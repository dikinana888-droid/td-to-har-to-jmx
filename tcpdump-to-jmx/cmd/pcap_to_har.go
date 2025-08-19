package cmd

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/tcpdump-to-jmx/pkg/converter"
)

var (
	pcapFile   string
	harOutput  string
	filterPort int
	filterHost string
)

var pcapToHarCmd = &cobra.Command{
	Use:   "pcap2har",
	Short: "Convert PCAP file to HAR format",
	Long:  `Convert a TCP dump (PCAP) file to HTTP Archive (HAR) format.`,
	RunE:  runPcapToHar,
}

func init() {
	pcapToHarCmd.Flags().StringVarP(&pcapFile, "input", "i", "", "Input PCAP file (required)")
	pcapToHarCmd.Flags().StringVarP(&harOutput, "output", "o", "", "Output HAR file")
	pcapToHarCmd.Flags().IntVarP(&filterPort, "port", "p", 0, "Filter by port (optional)")
	pcapToHarCmd.Flags().StringVarP(&filterHost, "host", "H", "", "Filter by host (optional)")
	
	pcapToHarCmd.MarkFlagRequired("input")
}

func runPcapToHar(cmd *cobra.Command, args []string) error {
	// Auto-generate output filename if not specified
	if harOutput == "" {
		base := strings.TrimSuffix(filepath.Base(pcapFile), filepath.Ext(pcapFile))
		harOutput = base + ".har"
	}

	printInfo("Converting PCAP to HAR...")
	printInfo("Input: %s", pcapFile)
	printInfo("Output: %s", harOutput)
	
	if filterPort > 0 {
		printInfo("Filtering by port: %d", filterPort)
	}
	if filterHost != "" {
		printInfo("Filtering by host: %s", filterHost)
	}

	converter := converter.NewPcapToHarConverter()
	
	// Set filters
	if filterPort > 0 {
		converter.SetPortFilter(filterPort)
	}
	if filterHost != "" {
		converter.SetHostFilter(filterHost)
	}

	err := converter.Convert(pcapFile, harOutput)
	if err != nil {
		printError("Failed to convert: %v", err)
		return err
	}

	printSuccess("Successfully converted PCAP to HAR")
	printSuccess("Output saved to: %s", harOutput)
	
	return nil
}