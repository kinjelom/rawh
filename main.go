package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"rawh/client"
	"rawh/common"
	"rawh/server"
	"strings"
)

var name = "rawh"
var description = "rawh functions either as an HTTP server or as a client to diagnose requests and responses."
var version = "dev"

func main() {
	var rootCmd = &cobra.Command{
		Use:   name,
		Short: description,
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			versionFlag, err := cmd.Flags().GetBool("version")
			if err == nil && versionFlag {
				fmt.Println(name, version)
				os.Exit(0)
			}
		},
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("No mode specified, running default client mode. Use 'rawh help' for more information.")
			_ = cmd.Help()
		},
	}
	rootCmd.PersistentFlags().BoolP("version", "V", false, "Displays the application version.")

	var verbose bool
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enables verbose output for the operation (client and server modes).")

	// Server commands
	var serverPort int
	var serverCmd = &cobra.Command{
		Use:   "server",
		Short: "Run as an HTTP server",
		Run: func(cmd *cobra.Command, args []string) {
			err := server.NewServer(serverPort, verbose).Serve()
			if err != nil {
				exitWithError(err)
			}
		},
	}
	serverCmd.Flags().IntVarP(&serverPort, "port", "p", 8080, "Specify the port the server will listen on")
	rootCmd.AddCommand(serverCmd)

	// Client commands
	var canonical bool
	var method string
	var httpVersionName string
	var tlsVersionName string
	var insecure bool
	var headers []string
	var normalizeHeaders bool
	var data string
	var generateDataSize string
	var clientCmd = &cobra.Command{
		Use:   "client <url>",
		Short: "Run as an HTTP client",
		Args:  cobra.ExactArgs(1), // Requires exactly one argument - url
		Run: func(cmd *cobra.Command, args []string) {

			var httpClient client.Client
			var err error
			if canonical {
				httpClient, err = client.NewCanonicalClient(tlsVersionName, insecure, httpVersionName, verbose)
			} else {
				httpClient, err = client.NewRawClient(normalizeHeaders, tlsVersionName, insecure, httpVersionName, verbose)
			}
			if err != nil {
				exitWithError(err)
			}
			if generateDataSize != "" {
				byteSize, err := common.ParsePrittyByteSize(generateDataSize)
				if err != nil {
					exitWithError(err)
				} else {
					data = common.GenerateSampleDataString(byteSize)
				}
			}
			url := args[0]
			if !strings.HasPrefix(url, "http") {
				url = "https://" + url
			}
			err = httpClient.DoRequest(method, url, headers, data)
			if err != nil {
				exitWithError(err)
			}
		},
	}
	clientCmd.Flags().BoolVarP(&canonical, "canonical", "C", false, "Specifies whether the 'canonical' client should be used; by default, the 'raw' client will be used.")
	clientCmd.Flags().StringVarP(&method, "method", "X", "GET", "Specifies the HTTP method to use (e.g., 'GET', 'POST').")
	clientCmd.Flags().StringVarP(&data, "data", "d", "", "Data to be sent as the body of the request, typically with 'POST'.")
	clientCmd.Flags().StringVar(&generateDataSize, "generate-data-size", "", "Data size [B|KB|MB|GB] to be generated and sent as the body of the request, typically with 'POST'.")
	clientCmd.Flags().StringVar(&httpVersionName, "http", "1.1", "Specifies the HTTP version to use (options: 1.0, 1.1, 2).")
	clientCmd.Flags().StringVar(&tlsVersionName, "tls", "1.2", "Specifies the TLS version to use (options: 1.0, 1.1, 1.2, 1.3).")
	clientCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure server connections.")
	clientCmd.Flags().StringArrayVarP(&headers, "header", "H", nil, "Adds a header to the request, format 'key: value'.")
	clientCmd.Flags().BoolVar(&normalizeHeaders, "normalize-headers", false, "Normalize header names format.")
	rootCmd.AddCommand(clientCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

}

func exitWithError(err error) {
	if err != nil {
		fmt.Println(name, version)
		_, _ = fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}
