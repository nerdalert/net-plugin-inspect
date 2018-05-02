//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
//
// This program inspects a Docker Networking Plugin.
//
// Syntax: inspectDockerNetworkingPlugin [options] docker-networking-plugin
//
//   Options:
//             [--docker-user]					Docker ID
//             [--docker-password]					Docker ID password
//			[--docker-registry-auth-endpoint]       Defaults to https://auth.docker.io
//             [--docker-registry-api-endpoint]        Defaults to https://registry-1.docker.io
//			[--test-script scriptname]              Specify an optional script to test the Docker Networking Plugin. The script gets passed 1 parameter - the Docker Networking Plugin name.
//             [--json]  						Generate Output in JSON to stdout
//			[--html]  						Generate Output in HTML
//             [-v]      						Verbose output
//             [-h]      						Help
//
// Pre-requisites:
//
//     Docker must be installed.
//

package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"log"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/docker/inspect_docker_image/dockerAPI"
	"github.com/howeyc/gopass"
)

var todaysDateTime = time.Now()

var stepNumber int
var jsonOutput = false
var htmlOutput = false
var exitCode int

var dockerNetworkingPluginTestScript string
var dockerNetworkingPluginGetLogsScript string

var dockerPluginManifest = dockerAPI.DockerImageManifest{}
var dockerPluginConfigurationBlob = dockerAPI.DockerPluginConfigurationBlob{}

var inspectionData = inspectionStruct{}

const termReportLineLength = 194
const termImageInformationLineLength = 164
const testNetworkName = "test_network"

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// This structure defines the structure to hold the Inspection Data and Results.
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type inspectionStruct struct {
	InspectionDate                             string
	SystemDockerVersion                        string
	SystemOperatingSystem                      string
	SystemArchitecture                         string
	Errors                                     int
	Warnings                                   int
	verboseOutput                              bool
	Messages                                   []string
	DockerNetworkingPlugin                     string
	DockerNetworkingPluginRepo                 string
	DockerNetworkingPluginTag                  string
	Description                                string
	InterfaceSocket                            string
	InterfaceSocketTypes                       string
	Documentation                              string
	DockerNetworkingPluginDockerVersion        string
	DockerNetworkingPluginDigest               string
	DockerNetworkingPluginBaseLayerImageDigest string
	EntryPoint                                 string
	WorkDir                                    string
	User                                       string
	IpcHost                                    string
	PidHost                                    string
	HTMLMessages                               []template.HTML
	TestResults                                []template.HTML
	HTMLReportFile                             string
	VulnerabilitiesScanURL                     string
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// This structure defines an array entry containing the JSON Output Test Results
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type jsonResultsStruct struct {
	Status  string
	Message string
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// This structure defines the JSON Output
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
type jsonOutputStruct struct {
	Date                                       string `json:"Date"`
	SystemOperatingSystem                      string `json:"SystemOperatingSystem"`
	SystemArchitecture                         string `json:"SystemArchitecture"`
	SystemDockerVersion                        string `json:"SystemDockerVersion"`
	DockerNetworkingPlugin                     string `json:"DockerLogginPlugin"`
	Description                                string `json:"Description"`
	Documentation                              string `json:"Documentation"`
	DockerNetworkingPluginDigest               string `json:"DockerNetworkingPluginDigest"`
	DockerNetworkingPluginBaseLayerImageDigest string `json:"BaseLayerImageDigest"`
	DockerNetworkingPluginDockerVersion        string `json:"DockerVersion,omitempty"`
	EntryPoint                                 string `json:"Entrypoint"`
	InterfaceSocket                            string `json:"InterfaceSocket"`
	InterfaceSocketTypes                       string `json:"InterfaceSocketTypes"`
	WorkDir                                    string `json:"WorkDir"`
	User                                       string `json:"User"`
	IpcHost                                    bool   `json:"IpcHost"`
	PidHost                                    bool   `json:"PidHost"`
	Errors                                     int    `json:"Errors"`
	Warnings                                   int    `json:"Warnings"`
	HTMLReportFile                             string `json:"HTMLReportFile"`
	VulnerabilitiesScanURL                     string
	Results                                    []jsonResultsStruct
}

var htmlTemplate = `<!DOCTYPE html>
<html>
<head>
<meta http-equiv='content-type' content='text/html; charset=utf-8' />
<meta name='author' content='Gary Forghetti' />
<meta name='copyright' content='Copyright 2017 Docker, Inc.' />
<style type=text/css>
body {
	background-color:white;
	font-weight: normal;
	color:black;
	padding-bottom:10px;
	padding-right:10px;
	padding-left:10px;
}
fieldset, legend {
	-moz-border-radius-bottomleft:7px;
	-moz-border-radius-bottomright:7px;
	-moz-border-radius-topleft:5px;
	-moz-border-radius-topright:7px;
	-webkit-border-radius:7px;
	background-color:inherit;
	border:3px solid black;
	border-radius:3px;
	color:inherit;
	display:block;
	font-size:inherit;
	font-family:inherit;
}
fieldset {
	margin-top:4px;
	margin-bottom:4px;
	margin-left:4px;
	margin-right:4px;
	padding-top:5px;
	padding-right:5px;
	padding-bottom:5px;
	padding-left:5px;
	width:auto;
}
legend {
	background-color:Gainsboro;
	color:black;
	font-size:larger;
	font-weight:bold;
	margin-top:4px;
	padding-top:2px;
	padding-right:3px;
	padding-bottom:3px;
	padding-left:3px;
}
table, th, td {
	border: 2px solid black;
	font-size:inherit;
	font-family:inherit;
}
table {
	background-color:inherit;
	border-collapse:collapse;
	border-spacing:2px;
	border-style:solid;
	border-width:thin;
	color:inherit;
	margin-top:4px;
	padding-right:3px;
	padding-left:3px;
	width:100%;
}
th, td {
	padding-top:5px;
	padding-right:5px;
	padding-bottom:5px;
	padding-left:5px;
	vertical-align:top;
	text-align:left;
}
th {
	background-color:rgb(0, 135, 201);
	color:white;
	font-weight:bold;
	white-space:nowrap;
	width:1px;
}
td {
	background-color:inherit;
	color:inherit;
	font-weight:inherit;
	white-space:inherit;
	width:auto;
}
a.doc:link, a.doc:visited, a.doc:active, a.doc:hover { text-decoration:underline;font-weight:bold;color:white; }
a.doc:hover { font-size:105%; }
a.ref:link, a.ref:visited, a.ref:active, a.ref:hover { text-decoration:underline;font-weight:bold;color:blue; }
a.ref:hover { font-size:105%; }
.success_message {
	background-color:green;
	color:white;
	font-weight:bold;
	padding-top:3px;
	padding-right:3px;
	padding-bottom:3px;
	padding-left:3px;
	white-space:nowrap;
	width:5%;
}
.warning_message {
	background-color:yellow;
	color:black;
	font-weight:bold;
	padding-top:3px;
	padding-right:3px;
	padding-bottom:3px;
	padding-left:3px;
	white-space:nowrap;
	width:5%;
}
.error_message {
	background-color:red;
	color:black;
	font-weight:bold;
	padding-top:3px;
	padding-right:3px;
	padding-bottom:3px;
	padding-left:3px;
	white-space:nowrap;
	width:5%;
}
</style>
<title>Docker networking plugin inspection report</title>
</head>
<body>
<h3><legend><div style='float:left;'>Docker networking plugin: <span style='color:blue;'>{{.DockerNetworkingPlugin}}</span></div>
<div style='float:right;'>Report Date: <span style='color:blue;'>{{.InspectionDate}}</span></div>
<div style='clear:both;'></div>
</legend></h3>
<fieldset>
<legend>Docker Plugin information</legend>
<table cols='2'>
<tr><th>Docker Plugin</th><td>{{.DockerNetworkingPlugin}}</td></tr>
<tr><th><a class='doc' href='https://docs.docker.com/engine/extend/config/' target='_blank'>Description</a></th><td>{{.Description}}</td></tr>
<tr><th><a class='doc' href='https://docs.docker.com/engine/extend/config/' target='_blank'>Documentation</a></th><td>{{.Documentation}}</td></tr>
<tr><th>Digest</th><td>{{.DockerNetworkingPluginDigest}}</td></tr>
<tr><th>Base layer digest</th><td>{{.DockerNetworkingPluginBaseLayerImageDigest}}</td></tr>
{{if .DockerNetworkingPluginDockerVersion}}<tr><th>Docker version</th><td>{{.DockerNetworkingPluginDockerVersion}}</td></tr>{{end}}
<tr><th><a class='doc' href='https://docs.docker.com/engine/extend/config/' target='_blank'>Interface Socket</a></th><td>{{.InterfaceSocket}}</td></tr>
<tr><th><a class='doc' href='https://docs.docker.com/engine/extend/config/' target='_blank'>Interface Socket Types</a></th><td>{{.InterfaceSocketTypes}}</td></tr>
<tr><th><a class='doc' href='https://docs.docker.com/engine/extend/config/' target='_blank'>IpcHost</a></th><td>{{.IpcHost}}</td></tr>
<tr><th><a class='doc' href='https://docs.docker.com/engine/extend/config/' target='_blank'>PidHost</a></th><td>{{.PidHost}}</td></tr>
<tr><th><a class='doc' href='https://docs.docker.com/engine/reference/builder/#entrypoint' target='_blank'>Entrypoint</a></th><td>{{.EntryPoint}}</td></tr>
<tr><th><a class='doc' href='https://docs.docker.com/engine/reference/builder/#workdir' target='_blank'>WorkDir</a></th><td>{{.WorkDir}}</td></tr>
<tr><th><a class='doc' href='https://docs.docker.com/engine/reference/builder/#user' target='_blank'>User</a></th><td>{{.User}}</td></tr>
</table>
</fieldset>
<br>
<br>
<fieldset>
<legend>Report Summary</legend>
<table cols='2'>
<tr><th>Date:</th><td>{{.InspectionDate}}</td></tr>
<tr><th>Operating system</th><td>{{.SystemOperatingSystem}}</td></tr>
<tr><th>Architecture</th><td>{{.SystemArchitecture}}</td></tr>
<tr><th>Docker Version</th><td>{{.SystemDockerVersion}}</td></tr>
</table>
<br>
<br>
<fieldset>
<legend>Inspection Results</legend>
<table cols='2'>
{{range .HTMLMessages}}
{{.}}
{{end}}
</table>
</fieldset>
{{if .VulnerabilitiesScanURL}}
<br>
<br>
<fieldset>
<legend>Security Scan Results</legend>
<a class='ref' href='{{.VulnerabilitiesScanURL}}' target='_blank'>Click here to view the Security Scan Results</a>
</fieldset>
<br>
{{end}}
<br>
<br>
<fieldset>
<legend>Reference documentation</legend>
<br><a class='ref' href='https://github.com/docker/cli/tree/master/docs/extend' target='_blank'>Networking driver plugins documentation</a>
<br><a class='ref' href='https://docs.docker.com/engine/extend/' target='_blank'>Docker Engine managed plugin system</a>
<br><a class='ref' href='https://docs.docker.com/engine/extend/plugins_network/' target='_blank'>Using a networking driver plugin</a>
</fieldset>
<br>
</fieldset>
<br>
</body>
</html>`

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Logs the Fatal error to JSON (if JSON output was requested) or stderr
// Do not call this function from either the generateHTMLReport() or generateJSONOutput() functions or it will be a recursive loop.
// Instead just call log.Fatal function.
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func logFatalError(err error) {
	if jsonOutput == true {
		printError(err.Error())
		generateJSONOutput()
	} else {
		log.Println(err)
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Prints a success message
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func printSuccess(message string) {
	inspectionData.HTMLMessages = append(inspectionData.HTMLMessages, formatHTMLSuccess(message))
	message = fmt.Sprintf("%-10s", "Passed:") + message
	printMessage(boldGreen(message))
	inspectionData.Messages = append(inspectionData.Messages, boldGreen(strings.Trim(message, "\n")))
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Prints a warning message
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func printWarning(message string) {
	inspectionData.HTMLMessages = append(inspectionData.HTMLMessages, formatHTMLWarning(message))
	message = fmt.Sprintf("%-10s", "Warning:") + message
	printMessage(boldYellow(message))
	inspectionData.Messages = append(inspectionData.Messages, boldYellow(strings.Trim(message, "\n")))
	inspectionData.Warnings++
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Prints an error message
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func printError(message string) {
	inspectionData.HTMLMessages = append(inspectionData.HTMLMessages, formatHTMLError(message))
	message = fmt.Sprintf("%-10s", "Error:") + message
	printMessage(boldRed(message))
	inspectionData.Messages = append(inspectionData.Messages, boldRed(strings.Trim(message, "\n")))
	inspectionData.Errors++
	exitCode = 1
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Prints a message to stdout only if JSON Output is not specified, because JSON output will be written to stdout
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func printMessage(message string) {
	if jsonOutput == false {
		fmt.Println(message)
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Returns an HTML table row containing the error message
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func formatHTMLError(message string) template.HTML {
	return template.HTML("<tr><td class='error_message'>Error</td><td>" + message + "</td></tr>")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Returns an HTML table row containing the warning message
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func formatHTMLWarning(message string) template.HTML {
	return template.HTML("<tr><td class='warning_message'>Warning</td><td>" + message + "</td></tr>")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Returns an HTML table row containing the success message
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func formatHTMLSuccess(message string) template.HTML {
	return template.HTML("<tr><td class='success_message'>Passed</td><td>" + message + "</td></tr>")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Returns an HTML table row containing the success message for verbose networking
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func formatHTMLVerboseSuccess(direction string, message string) template.HTML {
	return template.HTML("<tr><td class='success_message'>" + direction + "</td><td>" + message + "</td></tr>")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Returns an HTML table row containing the error message for verbose networking
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func formatHTMLVerboseError(direction string, message string) template.HTML {
	return template.HTML("<tr><td class='error_message'>" + direction + "</td><td>" + message + "</td></tr>")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Prints a verbose message containing log data
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func printVerbose(equal bool, direction string, message string) {
	if equal == true {
		inspectionData.HTMLMessages = append(inspectionData.HTMLMessages, formatHTMLVerboseSuccess(direction, message))
		message = fmt.Sprintf("%-10s", direction+":") + message
		printMessage(boldGreen(message))
		inspectionData.Messages = append(inspectionData.Messages, boldGreen(strings.Trim(message, "\n")))
	} else {
		inspectionData.HTMLMessages = append(inspectionData.HTMLMessages, formatHTMLVerboseError(direction, message))
		message = fmt.Sprintf("%-10s", direction+":") + message
		printMessage(boldRed(message))
		inspectionData.Messages = append(inspectionData.Messages, boldRed(strings.Trim(message, "\n")))
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Run a command and return the output and error
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func runCommand(command string) (string, error) {
	command = strings.TrimSpace(command)

	if command == "" {
		return "", fmt.Errorf("you must specify a command!")
	}

	var cmd string
	var args []string

	if runtime.GOOS == "windows" {
		cmd = "Powershell.exe"
		args = strings.Fields("-ExecutionPolicy Unrestricted -NoLogo -NoProfile -NonInteractive -Command")
		args = append(args, []string{command}...)
	} else {
		cmd = "/bin/bash"
		args = []string{"-c", command}
	}

	output, err := exec.Command(cmd, args...).CombinedOutput()
	return strings.TrimSpace(string(output)), err
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Returns a string which contains operating system information
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func getOperatingSystemInfo() (string, error) {
	if runtime.GOOS == "darwin" {
		output, err := runCommand(`defaults read loginwindow SystemVersionStampAsString`)
		if err != nil {
			fmt.Println(output)
			return "", err
		}
		return "Operating System: MacOS " + runtime.GOOS + " Version: " + output, nil
	} else if runtime.GOOS == "windows" {
		output, err := runCommand(`(Get-CimInstance win32_operatingsystem).caption`)
		if err != nil {
			fmt.Println(output)
			return "", err
		}
		return "Operating System: " + output, nil
	} else if runtime.GOOS == "linux" {
		output, err := runCommand(`cat /etc/os-release | grep 'PRETTY_NAME' | awk -F '"' '{ print $2 }'`)
		if err != nil {
			fmt.Println(output)
			return "", err
		}
		return "Operating System: " + output, nil
	}

	return "Operating System is unknown!", nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Format the passed message so it will be displayed in Bold Red and return it
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func boldRed(str string) string {
	if runtime.GOOS != "windows" && jsonOutput == false {
		return "\033[1;31m" + str + "\033[0m"
	}
	return str
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Format the passed message so it will be displayed in Bold Green and return it
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func boldGreen(str string) string {
	if runtime.GOOS != "windows" && jsonOutput == false {
		return "\033[1;32m" + str + "\033[0m"
	}
	return str
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Format the passed message so it will be displayed in Bold Yellow and return it
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func boldYellow(str string) string {
	if runtime.GOOS != "windows" && jsonOutput == false {
		return "\033[1;33m" + str + "\033[0m"
	}
	return str
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Print Report Header
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func printReportHeader(dockerNetworkingPlugin string) {
	printMessage("\n" + strings.Repeat("*", termReportLineLength))
	printMessage("* Docker Networking Plugin: " + dockerNetworkingPlugin)
	printMessage(strings.Repeat("*", termReportLineLength))
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Print Report Summary Header
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func printReportSummaryHeader(dockernetworkingPlugin string) {
	printMessage("\n" + strings.Repeat("*", termReportLineLength))
	printMessage("* Summary of the inspection for the Docker Networking Plugin: " + dockernetworkingPlugin)
	printMessage(strings.Repeat("*", termReportLineLength))
	printMessage("")
	printMessage("Report Date: " + inspectionData.InspectionDate)
	printMessage("Operating System: " + inspectionData.SystemOperatingSystem)
	printMessage("Architecture: " + inspectionData.SystemArchitecture)
	printMessage(inspectionData.SystemDockerVersion)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Print a step header surrounding the passed message
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func printStep(message string) {
	stepNumber++
	printMessage("\n" + strings.Repeat("*", termReportLineLength))
	printMessage(fmt.Sprintf("* Step #%d %s", stepNumber, message))
	printMessage(strings.Repeat("*", termReportLineLength))
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Truncates a string if it's length is greater than the specified length
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func truncateString(originalString string, maxSize int) string {
	if len(originalString) > maxSize {
		return originalString[:maxSize]
	}

	return originalString
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Parses an inspection result message at the first colon and returns the 2 parts separately
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func parseMessage(message string) (string, string) {
	index := strings.Index(message, ":")
	if index != -1 {
		status := message[:index]
		message := strings.TrimSpace(message[index+1:])
		return status, message
	}
	return "", message
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Generate the HTML report if HTML Output was requested
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func generateHTMLReport() {
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Create the html subdirectory if it does not exist
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	_, err := os.Stat(`html`)
	if os.IsNotExist(err) {
		errDir := os.MkdirAll(`html`, 0755)
		if errDir != nil {
			log.Fatal(err)
		}
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Open (over write) the HTML file
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	inspectionData.HTMLReportFile = `html/` + strings.Replace(inspectionData.DockerNetworkingPluginRepo, `/`, `-`, -1) + "-" + inspectionData.DockerNetworkingPluginTag +
		"_inspection_report_" + todaysDateTime.Format("2006-01-02_03-04-05") + ".html"
	file, err := os.Create(inspectionData.HTMLReportFile)
	if err != nil {
		log.Fatal(err)
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Create a defer function to close the HTML report file
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	defer func() {
		file.Close()
		if err != nil {
			log.Fatal(err)
		}
	}()

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Create a Template
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	t, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		log.Fatal(err)
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Write the HTML to the report html file using the template
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	err = t.Execute(file, inspectionData)
	if err != nil {
		log.Fatal(err)
	}

	printMessage(fmt.Sprintf("An HTML report has been generated in the file %s", inspectionData.HTMLReportFile))

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// If running on MacOS then open the report html file
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	if runtime.GOOS == "darwin" {
		runCommand("nohup open " + inspectionData.HTMLReportFile + " &")
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Generate JSON Output if JSON Output was requested
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func generateJSONOutput() {
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Create a new JSON Encoder
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	jsonEncoder := json.NewEncoder(os.Stdout)

	jsonOutputData := jsonOutputStruct{}

	jsonOutputData.Date = inspectionData.InspectionDate
	jsonOutputData.SystemOperatingSystem = inspectionData.SystemOperatingSystem
	jsonOutputData.SystemArchitecture = inspectionData.SystemArchitecture
	jsonOutputData.SystemDockerVersion = inspectionData.SystemDockerVersion
	jsonOutputData.DockerNetworkingPlugin = inspectionData.DockerNetworkingPlugin
	jsonOutputData.Description = inspectionData.Description
	jsonOutputData.Documentation = inspectionData.Documentation
	jsonOutputData.DockerNetworkingPluginDigest = inspectionData.DockerNetworkingPluginDigest
	jsonOutputData.DockerNetworkingPluginBaseLayerImageDigest = inspectionData.DockerNetworkingPluginBaseLayerImageDigest
	jsonOutputData.DockerNetworkingPluginDockerVersion = inspectionData.DockerNetworkingPluginDockerVersion
	jsonOutputData.InterfaceSocket = inspectionData.InterfaceSocket
	jsonOutputData.InterfaceSocketTypes = inspectionData.InterfaceSocketTypes
	jsonOutputData.IpcHost = dockerPluginConfigurationBlob.IpcHost
	jsonOutputData.PidHost = dockerPluginConfigurationBlob.PidHost
	jsonOutputData.EntryPoint = inspectionData.EntryPoint
	jsonOutputData.WorkDir = inspectionData.WorkDir
	jsonOutputData.User = inspectionData.User

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Grab the Inspection and Test Results
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	jsonOutputData.Errors = inspectionData.Errors
	jsonOutputData.Warnings = inspectionData.Warnings
	jsonOutputData.VulnerabilitiesScanURL = inspectionData.VulnerabilitiesScanURL
	if htmlOutput == true {
		jsonOutputData.HTMLReportFile = inspectionData.HTMLReportFile
	}
	for _, statusMessage := range inspectionData.Messages {
		status, message := parseMessage(statusMessage)
		jsonOutputData.Results = append(jsonOutputData.Results, jsonResultsStruct{status, message})
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Encode the Docker Official Images structure back into JSON and write it to the JSON file.
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	err := jsonEncoder.Encode(jsonOutputData)
	if err != nil {
		log.Fatal(err)
	}
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Display the Command Help
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func usage() {
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Inspects a Docker Networking Plugin to see if it conforms to best practices.")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Syntax: inspectDockerNetworkingPlugin [options] dockerNetworkingPlugin")
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "Options:")
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr)
	fmt.Fprintln(os.Stderr, "  dockerNetworkingPlugin\n\tThe Docker Networking Plugin to inspect. This argument is required.")
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Main Entry to the program
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func main() {
	var err error
	// Tests to be run
	// pluginName := "weaveworks/net-plugin:latest_release"
	// docker plugin install weaveworks/net-plugin:latest_release
	// docker plugin disable weaveworks/net-plugin:latest_release
	// docker plugin rm weaveworks/net-plugin:latest_release
	// docker plugin ls
	// docker swarm init
	// docker network create --driver=weaveworks/net-plugin:latest_release mynetwork
	// docker network rm mynetwork

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Initialize some of the report data in the template
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	inspectionData.InspectionDate = todaysDateTime.Format("Mon Jan 02 15:04:05 2006")
	inspectionData.SystemArchitecture = runtime.GOARCH
	inspectionData.SystemOperatingSystem, err = getOperatingSystemInfo()
	if err != nil {
		logFatalError(err)
		os.Exit(1)
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Get the DOCKER_REGISTRY_AUTH_ENDPOINT Environment variable
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	var dockerRegistryAuthEndpoint = os.Getenv("DOCKER_REGISTRY_AUTH_ENDPOINT")
	if dockerRegistryAuthEndpoint == "" {
		dockerRegistryAuthEndpoint = "https://auth.docker.io"
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Get the DOCKER_REGISTRY_API_ENDPOINT Environment variable
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	var dockerRegistryAPIEndpoint = os.Getenv("DOCKER_REGISTRY_API_ENDPOINT")
	if dockerRegistryAPIEndpoint == "" {
		dockerRegistryAPIEndpoint = "https://registry-1.docker.io"
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Set the Log Flags to display the short file name and line number when Networking error messages using the log Package
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Setup, parse and verify the command line options
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	dockerUserIDPtr := flag.String("docker-user", "", " Docker User ID.  This overrides the DOCKER_USER environment variable.")
	dockerPasswordPtr := flag.String("docker-password", "", " Docker Password.  This overrides the DOCKER_PASSWORD environment variable.")
	dockerRegistryAuthEndpointPtr := flag.String("docker-registry-auth-endpoint", dockerRegistryAuthEndpoint, " Docker Registry Authentication Endpoint. "+
		"This overrides the DOCKER_REGISTRY_AUTH_ENDPOINT environment variable.")
	dockerRegistryAPIEndpointPtr := flag.String("docker-registry-api-endpoint", dockerRegistryAPIEndpoint, " Docker Registry API Endpoint. "+
		"This overrides the DOCKER_REGISTRY_API_ENDPOINT environment variable.")
	jsonPtr := flag.Bool("json", false, " Generate JSON output.")
	htmlPtr := flag.Bool("html", false, " Generate HTML output.")
	helpPtr := flag.Bool("help", false, " Help on the command.")
	verbosePtr := flag.Bool("verbose", false, " Displays more verbose output.")

	flag.Usage = usage
	flag.Parse()

	dockerAPI.DockerRegistryAuthEndpoint = *dockerRegistryAuthEndpointPtr
	dockerAPI.DockerRegistryAPIEndpoint = *dockerRegistryAPIEndpointPtr
	jsonOutput = *jsonPtr
	htmlOutput = *htmlPtr
	inspectionData.verboseOutput = *verbosePtr

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Display the command usage if the help command line option was specified
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	if *helpPtr == true {
		usage()
		os.Exit(0)
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Check for the 1 command argument which is the name of the Docker Networking Plugin to inspect
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	if flag.NArg() == 0 {
		usage()
		os.Exit(1)
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Get the Docker Networking Plugin and parse it into the repo and tag
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	inspectionData.DockerNetworkingPlugin = flag.Arg(0)

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Make sure nothing is appended to the DockerNetworkingPlugin that might cause security issues
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	re1 := regexp.MustCompile(`^(.*?)[;|&].*$`)
	if match := re1.FindStringSubmatch(inspectionData.DockerNetworkingPlugin); len(match) != 0 {
		inspectionData.DockerNetworkingPlugin = string(match[1])
	}

	if strings.Index(inspectionData.DockerNetworkingPlugin, "/") == -1 {
		logFatalError(errors.New("you did not prefix the Docker Networking Plugin with a user name (username/, library/ or dockerstorestaging/)!"))
		os.Exit(1)
	}

	var tagIndex = -1
	if tagIndex = strings.LastIndex(inspectionData.DockerNetworkingPlugin, ":"); tagIndex == -1 {
		logFatalError(errors.New("the Docker Networking Plugin does not contain a tag!"))
		os.Exit(1)
	}

	inspectionData.DockerNetworkingPluginRepo = inspectionData.DockerNetworkingPlugin[:tagIndex]
	inspectionData.DockerNetworkingPluginTag = inspectionData.DockerNetworkingPlugin[tagIndex+1:]
	if inspectionData.DockerNetworkingPluginTag == "" {
		logFatalError(errors.New("the Docker Networking Plugin does not contain a tag!"))
		os.Exit(1)
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Get the Docker User ID from the command parameter. If "blank" then get the DOCKER_USER environment variable, otherwise prompt the user.
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	dockerUser := *dockerUserIDPtr
	if dockerUser == "" {
		dockerUser = os.Getenv("DOCKER_USER")
	}

	for dockerUser == "" {
		fmt.Print("Enter your Docker User ID: ")
		fmt.Scanf("%s\n", &dockerUser)
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Get the Docker Password from the command parameter. If "blank" then get the DOCKER_PASSWORD environment variable, otherwise prompt the user.
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	dockerPassword := *dockerPasswordPtr
	if dockerPassword == "" {
		dockerPassword = os.Getenv("DOCKER_PASSWORD")
	}

	for dockerPassword == "" {
		fmt.Print("Enter your Docker Password: ")
		pass, err := gopass.GetPasswdMasked()
		if err != nil {
			logFatalError(err)
		}

		dockerPassword = string(pass)
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Login to Docker
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	output, err := runCommand(fmt.Sprintf("docker login --username %s --password %s %s", dockerUser, dockerPassword, dockerRegistryAPIEndpoint))
	if err != nil {
		logFatalError(errors.New(err.Error() + "\n" + output))
		os.Exit(1)
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Get the Docker Version
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	inspectionData.SystemDockerVersion, err = runCommand("docker --version")
	if err != nil {
		logFatalError(errors.New(err.Error() + "\n" + inspectionData.SystemDockerVersion))
		os.Exit(1)
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Print the Docker Networking Plugin inspection report header
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	printReportHeader(inspectionData.DockerNetworkingPlugin)

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Inspecting the Docker Networking Plugin
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	printStep("Inspecting the Docker Networking Plugin: " + inspectionData.DockerNetworkingPlugin + " ...")

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Get the Docker Networking Plugin image Digest for the Docker Plugin
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////

	inspectionData.DockerNetworkingPluginDigest, err = dockerAPI.GetDockerImageDigest(dockerUser, dockerPassword, inspectionData.DockerNetworkingPlugin, `plugin`)
	if err != nil {
		logFatalError(err)
		os.Exit(1)
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Get the Docker Networking Plugin image Manifest for the Docker Networking Plugin
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	dockerPluginManifest = dockerAPI.DockerImageManifest{}
	err = dockerAPI.GetDockerImageManifest(dockerUser, dockerPassword, inspectionData.DockerNetworkingPlugin, `plugin`, &dockerPluginManifest)
	if err != nil {
		logFatalError(err)
		os.Exit(1)
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Get the Docker Docker Networking Plugin image's base layer image digest
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	inspectionData.DockerNetworkingPluginBaseLayerImageDigest = dockerPluginManifest.Layers[0].Digest

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Get the Docker Configuration Blob for the Docker Networking Plugin
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	dockerPluginConfigurationBlob = dockerAPI.DockerPluginConfigurationBlob{}
	err = dockerAPI.GetDockerPluginConfigBlob(dockerUser, dockerPassword, inspectionData.DockerNetworkingPlugin, &dockerPluginConfigurationBlob)
	if err != nil {
		logFatalError(err)
		os.Exit(1)
	}

	successMessage := fmt.Sprintf("Docker Networking Plugin image %s has been inspected.", inspectionData.DockerNetworkingPlugin)
	printSuccess(successMessage)

	inspectionData.DockerNetworkingPluginDockerVersion = dockerPluginConfigurationBlob.DockerVersion
	inspectionData.Description = dockerPluginConfigurationBlob.Description
	inspectionData.Documentation = dockerPluginConfigurationBlob.Documentation
	inspectionData.InterfaceSocket = dockerPluginConfigurationBlob.Interface.Socket
	inspectionData.InterfaceSocketTypes = strings.Join(dockerPluginConfigurationBlob.Interface.Types, "")
	inspectionData.IpcHost = strconv.FormatBool(dockerPluginConfigurationBlob.IpcHost)
	inspectionData.PidHost = strconv.FormatBool(dockerPluginConfigurationBlob.PidHost)
	inspectionData.WorkDir = dockerPluginConfigurationBlob.WorkDir
	inspectionData.EntryPoint = strings.Join(dockerPluginConfigurationBlob.Entrypoint, " ")
	for user := range dockerPluginConfigurationBlob.User {
		inspectionData.User += user + " "
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Print the Docker Networking Plugin Information
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	printStep("Docker Networking Plugin information")

	var lineFormat = fmt.Sprintf("| %%-23s | %%-%ds |", termImageInformationLineLength)
	var separator = "+" + strings.Repeat("-", 25) + "+" + strings.Repeat("-", termImageInformationLineLength+2) + "+"

	printMessage(separator)
	printMessage(fmt.Sprintf(lineFormat, "Docker Networking Plugin:", inspectionData.DockerNetworkingPlugin))
	printMessage(fmt.Sprintf(lineFormat, "Description:", inspectionData.Description))
	printMessage(fmt.Sprintf(lineFormat, "Documentation:", inspectionData.Documentation))
	printMessage(fmt.Sprintf(lineFormat, "Digest:", inspectionData.DockerNetworkingPluginDigest))
	printMessage(fmt.Sprintf(lineFormat, "Base layer digest:", inspectionData.DockerNetworkingPluginBaseLayerImageDigest))

	if inspectionData.DockerNetworkingPluginDockerVersion != "" {
		printMessage(fmt.Sprintf(lineFormat, "Docker version:", inspectionData.DockerNetworkingPluginDockerVersion))
	}

	printMessage(fmt.Sprintf(lineFormat, "Interface Socket:", inspectionData.InterfaceSocket))
	printMessage(fmt.Sprintf(lineFormat, "Interface Socket Types:", inspectionData.InterfaceSocketTypes))
	printMessage(fmt.Sprintf(lineFormat, "IpcHost:", inspectionData.IpcHost))
	printMessage(fmt.Sprintf(lineFormat, "PidHost:", inspectionData.PidHost))
	printMessage(fmt.Sprintf(lineFormat, "Entrypoint:", truncateString(inspectionData.EntryPoint, termImageInformationLineLength)))
	printMessage(fmt.Sprintf(lineFormat, "WorkDir:", inspectionData.WorkDir))
	printMessage(fmt.Sprintf(lineFormat, "User:", inspectionData.User))
	printMessage(separator)

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Install the Docker Networking Plugin		/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	if installDockerNetworkingPlugin(inspectionData.DockerNetworkingPlugin) {
		////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		// Now run the Networking Plugin Tests
		////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		//runNetworkingPluginTest()
		runNetworkingPluginTest(inspectionData.DockerNetworkingPlugin)
		//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		// Remove the Docker Networking Plugin if it was installed
		//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
		printStep("Removing the Docker networking plugin")
		removeDockerNetworkingPlugin(inspectionData.DockerNetworkingPlugin)
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Print the Summary of the Docker Networking Plugin inspection
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	printReportSummaryHeader(inspectionData.DockerNetworkingPlugin)

	printMessage("")

	if inspectionData.Warnings > 0 {
		printMessage(fmt.Sprintf("There were %d %s detected!", inspectionData.Warnings, boldYellow("warnings")))
	}

	if inspectionData.Errors > 0 {
		printMessage(fmt.Sprintf("There were %d %s detected!", inspectionData.Errors, boldRed("errors")))
	}

	printMessage("\n" + strings.Join(inspectionData.Messages, "\n") + "\n")

	printMessage(fmt.Sprintf("The inspection of the Docker networking plugin %s has completed.", inspectionData.DockerNetworkingPlugin))

	// If the Docker Networking plugin being inspection is in dockertorestaging then display the vulnerabilities scan URL

	// If the Docker Image being inspected is in dockerstorestaging then display the vulnerabilities scan URL
	if strings.HasPrefix(inspectionData.DockerNetworkingPluginRepo, "dockerstorestaging/") {
		inspectionData.VulnerabilitiesScanURL = fmt.Sprintf("https://cloud.docker.com/app/dockerstorestaging/repository/docker/%s/tags/%s", inspectionData.DockerNetworkingPluginRepo, inspectionData.DockerNetworkingPluginTag)
		printMessage(fmt.Sprintf("\nVulnerabilities scan report URL: %s", inspectionData.VulnerabilitiesScanURL))
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Generate the HTML Report if HTML Output was requested
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	if htmlOutput == true {
		generateHTMLReport()
	}

	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Generate JSON Output if JSON Output was requested
	/////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	if jsonOutput == true {
		generateJSONOutput()
	}

	printMessage("")

	os.Exit(exitCode)
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Install the Docker Networking Plugin
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func installDockerNetworkingPlugin(dockerNetworkingPlugin string) bool {
	var output string
	var err error

	runCommand("docker swarm init")

	printStep(fmt.Sprintf("Installing the Docker Networking plugin %s ...", inspectionData.DockerNetworkingPlugin))

	if dockerNetworkingPluginInstalled(dockerNetworkingPlugin) {
		printWarning(fmt.Sprintf("The Docker networking plugin %s is already installed and will be removed.", inspectionData.DockerNetworkingPlugin))
		removeDockerNetworkingPlugin(dockerNetworkingPlugin)
	}

	output, err = runCommand("docker plugin install --grant-all-permissions " + dockerNetworkingPlugin)
	if err != nil {
		var errMessage = "Unable to install the Docker Networking Plugin!"
		if output != "" {
			errMessage = errMessage + ", " + output
		} else {
			errMessage = errMessage + ", " + err.Error()
		}
		printError(errMessage)
		return false
	}

	printSuccess(fmt.Sprintf("Docker networking plugin %s has been installed successfully.", inspectionData.DockerNetworkingPlugin))
	return true
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Removes the Docker Networking Plugin
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func removeDockerNetworkingPlugin(pluginName string) bool {
	output, err := runCommand("docker plugin remove " + pluginName + " --force")
	if err != nil {
		var errMessage = fmt.Sprintf("Unable to remove the Docker networking plugin %s!", pluginName)
		if output != "" {
			errMessage = errMessage + ", " + output
		} else {
			errMessage = errMessage + ", " + err.Error()
		}
		printError(errMessage)
		return false
	}

	printSuccess(fmt.Sprintf("Docker network plugin %s was removed.", pluginName))
	return true
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Creates the Docker Network
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func createDockerNetwork(pluginName string) bool {

	runCommand("docker network rm test_network")
	time.Sleep(1 * time.Second)
	output, err := runCommand("docker network create --driver=" + pluginName + " " + testNetworkName)
	if err != nil {
		var errMessage = fmt.Sprintf("Unable to create a Docker network using plugin %s!", pluginName)
		if output != "" {
			errMessage = errMessage + ", " + output
		} else {
			errMessage = errMessage + ", " + err.Error()
		}
		printError(errMessage)
		return false
	}
	time.Sleep(1 * time.Second)
	printSuccess(fmt.Sprintf("Docker network was created using plugin %s", pluginName))

	return true
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Removes the Docker Test Network
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func removeDockerNetwork(pluginName string) bool {

	time.Sleep(1 * time.Second)
	output, err := runCommand("docker network rm " + testNetworkName)
	if err != nil {
		var errMessage = fmt.Sprintf("Unable to remove the Docker test network using plugin %s!", pluginName)
		if output != "" {
			errMessage = errMessage + ", " + output
		} else {
			errMessage = errMessage + ", " + err.Error()
		}
		printError(errMessage)
		return false
	}

	printSuccess(fmt.Sprintf("Docker network was removed using plugin %s", pluginName))
	return true
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Checks to see if the Docker Networking Plugin is installed
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func dockerNetworkingPluginInstalled(dockerNetworkPlugin string) bool {
	_, err := runCommand("docker plugin inspect --format '{{ .Name }}' " + dockerNetworkPlugin)
	if err != nil {
		return false
	}

	return true
}

//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
// Test the Docker Networking Plugin
//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
func runNetworkingPluginTest(pluginName string) {
	//var err error
	//var containerID string

	//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	// Test the Docker Networking Plugin
	//////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////////
	printStep("Testing the Docker network creation using plugin: " + pluginName + " ...")

	if ok := createDockerNetwork(pluginName); !ok {
		printError("Docker Network Plugin Test has failed! Unable to create a Docker network using the plugin: " + pluginName)
	}

	printStep("Testing the Docker network deletion using plugin: " + pluginName + " ...")

	if ok := removeDockerNetwork(pluginName); !ok {
		printError("Docker Network Plugin Test has failed! Unable to delete a Docker network using the plugin: " + pluginName)
	}
}
