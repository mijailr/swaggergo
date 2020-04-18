package main

import (
	"bytes"
	"flag"
	"fmt"
	"github.com/oleiade/reflections"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"reflect"
	"strings"
	"time"
)

const swaggerHubUrl = "https://api.swaggerhub.com/apis"

var commandLineName = "swaggergo"
var commandLineVersion = "1.0.0"
var commandLineUsage = `swaggergo is an utility for publishing OpenAPI definitions to SwaggerHub.

Usage:
  $ swaggergo path/to/openapi.yml --type (yml | json) --oas 3.0.0 --api mijailr/sample-api --access-token [...]

Environment variables can also be used:

  $ export SWAGGERHUB_ACCESS_TOKEN="..."
  $ export SWAGGERHUB_API="..."
  $ swaggergo --file path/to/openapi.yml --type (yml | json)

Version:
  $ swaggergo --version

Help:
  $ swaggergo --help

See https://github.com/mijailr/swaggergo for more information.`

type commandLineOptions struct {
	SwaggerHubAccessToken string `flag:"access-token" env:"SWAGGERHUB_ACCESS_TOKEN" required:"true"`
	SwaggerHubApi         string `flag:"api" env:"SWAGGERHUB_API" required:"true"`
	Type                  string `flag:"type" default:"yml"`
	Oas                   string `flag:"oas" default:"3.0.0"`
}

func main() {
	if len(os.Args) == 1 {
		exitAndError("invalid usage")
	}

	if os.Args[1] == "--version" {
		fmt.Printf("%s version %s\n", commandLineName, commandLineVersion)
		os.Exit(0)
	}

	openApiFile := os.Args[1]
	if strings.HasPrefix(openApiFile, "--") {
		exitAndError("invalid usage")
	}

	options := commandLineOptions{}
	parseArgs(&options, os.Args)

	publish(openApiFile, &options)
}

func parseArgs(opts *commandLineOptions, args []string) {
	flags := flag.NewFlagSet(commandLineName, flag.ExitOnError)
	fields, _ := reflections.Fields(opts)

	for i := 0; i < len(fields); i++ {
		fieldName := fields[i]
		flagName, _ := reflections.GetFieldTag(opts, fieldName, "flag")
		fieldKind, _ := reflections.GetFieldKind(opts, fieldName)

		if fieldKind == reflect.String {
			flags.String(flagName, "", "")
		} else if fieldKind == reflect.Bool {
			flags.Bool(flagName, false, "")
		} else {
			exitAndError(fmt.Sprintf("Could not create flag for %s", fieldName))
		}
	}

	flags.Usage = func() {
		fmt.Printf("%s\n", commandLineUsage)
	}

	var argumentFlags []string
	started := false
	for i := 0; i < len(args); i++ {
		if strings.HasPrefix(args[i], "--") {
			started = true
		}

		if started {
			argumentFlags = append(argumentFlags, args[i])
		}
	}

	flags.Parse(argumentFlags)

	for i := 0; i < len(fields); i++ {
		fieldName := fields[i]
		fieldKind, _ := reflections.GetFieldKind(opts, fieldName)

		flagName, _ := reflections.GetFieldTag(opts, fieldName, "flag")
		envName, _ := reflections.GetFieldTag(opts, fieldName, "env")
		required, _ := reflections.GetFieldTag(opts, fieldName, "required")
		defaultValue, _ := reflections.GetFieldTag(opts, fieldName, "default")

		value := flags.Lookup(flagName).Value.String()

		if value == "" {
			value = os.Getenv(envName)
		}

		if required == "true" && value == "" {
			exitAndError(fmt.Sprintf("missing %s", flagName))
		}

		if value == "" && defaultValue != "" {
			value = defaultValue
		}

		if fieldKind == reflect.String {
			reflections.SetField(opts, fieldName, value)
		} else if fieldKind == reflect.Bool {
			// The bool is converted to a string above
			if value == "true" {
				reflections.SetField(opts, fieldName, true)
			} else {
				reflections.SetField(opts, fieldName, false)
			}
		} else {
			exitAndError(fmt.Sprintf("Could not set value of %s", fieldName))
		}
	}
}

func exitAndError(message interface{}) {
	fmt.Printf("%s: %s\nSee '%s --help'\n", commandLineName, message, commandLineName)
	os.Exit(1)
}

func publish(openApiPath string, options *commandLineOptions) {
	log.Printf("Creating release %s for repository: %s", openApiPath, options.SwaggerHubApi)

	repositoryParts := strings.Split(options.SwaggerHubApi, "/")
	if len(repositoryParts) != 2 {
		exitAndError("api is in the wrong format")
	}

	openApi, err := ioutil.ReadFile(openApiPath)
	if err != nil {
		exitAndError(fmt.Sprintf("can't read the file %s", openApiPath))
	}

	mediaType := "application/yaml"
	if options.Type == "json" {
		mediaType = "application/json"
	}

	response, err := postToSwaggerHub(openApi, mediaType, options)
	if err != nil {
		exitAndError("problem connecting to swaggerhub")
	}

	log.Printf("OpenApi sended with response: %s", response)
}

func postToSwaggerHub(openApi []byte, mediaType string, options *commandLineOptions) (response string, err error) {
	apiUrl := fmt.Sprintf("%s/%s?oas=%s", swaggerHubUrl, options.SwaggerHubApi, options.Oas)
	request, _ := http.NewRequest("POST", apiUrl, bytes.NewBuffer(openApi))
	request.Header.Set("Authorization", options.SwaggerHubAccessToken)
	request.Header.Set("accept", "application/json")
	request.Header.Set("Content-Type", mediaType)

	log.Printf("sending requiest to: %s", apiUrl)

	client := client()
	resp, err := client.Do(request)
	if err != nil {
		return "", err
	}
	bodyBytes, err := ioutil.ReadAll(request.Body)
	if err != nil {
		log.Fatal(err)
	}
	bodyString := string(bodyBytes)
	log.Print(bodyString)

	return resp.Status, nil
}

func client() http.Client {
	timeount := 10 * time.Second
	return http.Client{
		Timeout: timeount,
	}
}
