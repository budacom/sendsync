/*
Copyright © 2021 NAME HERE <EMAIL ADDRESS>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"strings"

	"os"

	"github.com/sendgrid/sendgrid-go"
	"github.com/spf13/cobra"
)

type template struct {
	Generation string
	Id         string
	Name       string
	UpdatedAt  string `json:"updated_at"`
	Versions   []version
}

type version struct {
	Active               int
	Editor               string
	Id                   string
	Name                 string
	TemplateId           string `json:"template_id"`
	ThumbnailUrl         string `json:"thumbnail_url"`
	UpdatedAt            string `json:"updated_at"`
	HtmlContent          string `json:"html_content"`
	PlainContent         string `json:"plain_content"`
	Subject              string
	GeneratePlainContent bool `json:"generate_plain_content"`
}

type templates struct {
	Templates []template
}

var apiKey = os.Getenv("SENDGRID_API_KEY")
var host = "https://api.sendgrid.com"

// templateCmd represents the get command
var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "A brief description of your command template",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("get called template")
		fmt.Println("Here are the arguments of card command : " + strings.Join(args, ","))
		templates := getTemplates()
		makeDir("templates")
		for _, template := range templates.Templates {
			fmt.Println(template.Name)
			dirPath := fmt.Sprintf("templates/%s", template.Name)
			makeDir(dirPath)
			templateJson, err := json.MarshalIndent(template, "", "  ")
			if err != nil {
				fmt.Println(err)
			}
			writeFile(fmt.Sprintf("%s/template.json", dirPath), string(templateJson))
			updateTemplate(template)
			activeVersion, err := findActiveVersion(template)
			if err != nil {
				fmt.Println(err)
			}
			writeFile(fmt.Sprintf("%s/content.html", dirPath), activeVersion.HtmlContent)
			writeFile(fmt.Sprintf("%s/content.txt", dirPath), activeVersion.PlainContent)
		}
	},
}

func findActiveVersion(template template) (version, error) {
	for _, version := range template.Versions {
		if version.Active == 1 {
			return version, nil
		}
	}
	return version{}, errors.New("active version not found")
}

func writeFile(path string, content string) {
	err := ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		log.Println(err)
	}
}

func updateTemplate(template template) {
	requestPath := fmt.Sprintf("/v3/templates/%s", template.Id)
	request := sendgrid.GetRequest(apiKey, requestPath, host)
	request.Method = "GET"
	response, err := sendgrid.API(request)
	if err != nil {
		log.Println(err)
	}
	err = json.Unmarshal([]byte(response.Body), &template)
	if err != nil {
		panic(err)
	}
}

func makeDir(path string) {
	err := os.Mkdir(path, 0755)
	if err != nil {
		fmt.Println(err)
	}
}

func getTemplates() templates {
	request := sendgrid.GetRequest(apiKey, "/v3/templates", host)
	request.Method = "GET"
	queryParams := make(map[string]string)
	queryParams["generations"] = "dynamic"
	request.QueryParams = queryParams
	response, err := sendgrid.API(request)
	if err != nil {
		log.Println(err)
	}

	var jsonMap templates
	err = json.Unmarshal([]byte(response.Body), &jsonMap)
	if err != nil {
		panic(err)
	}
	return jsonMap
}

func init() {
	getCmd.AddCommand(templateCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// getCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// getCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
