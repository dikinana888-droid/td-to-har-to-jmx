package cmd

import (
	"fmt"
	"os"

	"github.com/fatih/color"
	"github.com/spf13/cobra"
)

var (
	Version = "1.0.0"
	rootCmd = &cobra.Command{
		Use:   "tcpdump-to-jmx",
		Short: "Convert TCP dump to HAR and JMX formats",
		Long: `A tool for converting TCP dump (PCAP) files to HAR format and then to JMX format
for Apache JMeter with automatic correlation and parameterization.`,
		Version: Version,
	}
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.AddCommand(pcapToHarCmd)
	rootCmd.AddCommand(harToJmxCmd)
	rootCmd.AddCommand(convertCmd)
	
	// Set custom version template
	rootCmd.SetVersionTemplate(fmt.Sprintf("%s {{.Version}}\n", 
		color.GreenString("tcpdump-to-jmx version")))
}

func printError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, color.RedString("Error: "+format+"\n"), args...)
}

func printSuccess(format string, args ...interface{}) {
	fmt.Printf(color.GreenString("✓ "+format+"\n"), args...)
}

func printInfo(format string, args ...interface{}) {
	fmt.Printf(color.CyanString("ℹ "+format+"\n"), args...)
}