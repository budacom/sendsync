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
	"path/filepath"

	"github.com/sendgrid/sendgrid-go"
	"github.com/spf13/cobra"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("apply called")
		file, _ := cmd.Flags().GetString("file")
		fmt.Println("value of the flag name :" + file)

		dir := filepath.Dir(file)
		template := getTemplateFromFile(file)
		html := readFile(filepath.Join(dir, "content.html"))
		plain := readFile(filepath.Join(dir, "content.txt"))
		// vamos a sendgrid

		targetTemplate, err := getTemplateByName(template.Name)
		if err != nil {
			fmt.Println(err)
		}

		activeVersion, err := findActiveVersion(targetTemplate)
		if err != nil {
			fmt.Println(err)
		}
		activeVersion.HtmlContent = html
		activeVersion.PlainContent = plain

		requestUri := fmt.Sprintf("/v3/templates/%s/versions/%s", activeVersion.TemplateId, activeVersion.Id)
		request := sendgrid.GetRequest(apiKey, requestUri, host)
		request.Method = "PATCH"

		versionJson, err := json.Marshal(activeVersion)
		if err != nil {
			fmt.Println(err)
		}

		request.Body = versionJson
		response, err := sendgrid.API(request)
		if err != nil {
			log.Println(err)
		} else {
			fmt.Println(response.StatusCode)
			fmt.Println(response.Body)
			fmt.Println(response.Headers)
		}
	},
}

func readFile(file string) string {
	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return string(fileContent)
}

func getTemplateFromFile(file string) template {
	var templateFromFile template
	templateJson := readFile(file)
	err := json.Unmarshal([]byte(templateJson), &templateFromFile)
	if err != nil {
		panic(err)
	}
	return templateFromFile
}

func getTemplateByName(name string) (template, error) {
	templates := getTemplates()
	for _, template := range templates.Templates {
		if template.Name == name {
			return template, nil
		}
	}
	return template{}, errors.New("no template found for name")
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.PersistentFlags().StringP("file", "f", "", "Template manifest to apply")
	applyCmd.MarkPersistentFlagRequired("file")

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// applyCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// applyCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
