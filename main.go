package main

import (
	"fmt"
	"github.com/spf13/cobra"
	"os"
	"rawh/client"
	"rawh/common"
	"rawh/server"
)

var name = "rawh"
var description = "rawh functions either as an HTTP server or as a client to diagnose requests and responses."
var version = "dev"

func main() {
	var verbose bool
	var showVersion bool
	var serverPort int
	var rootCmd = &cobra.Command{
		Use:   name,
		Short: description,
		Run: func(cmd *cobra.Command, args []string) {
			if showVersion {
				fmt.Println(name, version)
				os.Exit(0)
			}
		},
	}
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "Enables verbose output for the operation (client and server modes).")
	rootCmd.PersistentFlags().BoolVarP(&showVersion, "version", "V", false, "Displays the application version.")

	// Server commands
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
	var method string
	var url string
	var httpVersionName string
	var tlsVersionName string
	var insecure bool
	var headers []string
	var normalizeHeaders bool
	var data string
	var generateDataSize int
	var clientCmd = &cobra.Command{
		Use:   "client",
		Short: "Run as an HTTP client",
		Run: func(cmd *cobra.Command, args []string) {
			newClient, err := client.NewClient(normalizeHeaders, tlsVersionName, insecure, httpVersionName, verbose)
			if err != nil {
				exitWithError(err)
			}
			if generateDataSize > 0 {
				data = common.GenerateSampleDataString(generateDataSize)
			}
			err = newClient.DoRequest(method, url, headers, data)
			if err != nil {
				exitWithError(err)
			}
		},
	}
	clientCmd.Flags().StringVar(&url, "url", "http://localhost:8080", "Specifies the URL for the client request")
	clientCmd.Flags().StringVarP(&method, "method", "X", "GET", "Specifies the HTTP method to use (e.g., 'GET', 'POST').")
	clientCmd.Flags().StringVarP(&data, "data", "d", "", "Data to be sent as the body of the request, typically with 'POST'.")
	clientCmd.Flags().IntVar(&generateDataSize, "generate-data-size", 0, "Data size [bytes] to be generated and sent as the body of the request, typically with 'POST'.")
	clientCmd.Flags().StringVar(&httpVersionName, "http", "1.1", "Specifies the HTTP version to use (options: 1.0, 1.1, 2).")
	clientCmd.Flags().StringVar(&tlsVersionName, "tls", "1.2", "Specifies the TLS version to use (options: 1.0, 1.1, 1.2, 1.3).")
	clientCmd.Flags().BoolVarP(&insecure, "insecure", "k", false, "Allow insecure server connections.")
	clientCmd.Flags().StringArrayVarP(&headers, "header", "H", nil, "Adds a header to the request, format 'key: value'.")
	clientCmd.Flags().BoolVar(&normalizeHeaders, "normalize-headers", false, "Normalize header names format.")
	rootCmd.AddCommand(clientCmd)

	// Default behavior if no subcommand is provided
	rootCmd.Run = func(cmd *cobra.Command, args []string) {
		fmt.Println("No mode specified, running default client mode. Use 'rawh help' for more information.")
		err := cmd.Help()
		if err != nil {
			exitWithError(err)
		}
	}

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
