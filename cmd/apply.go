/*
Copyright © 2021 Buda.com

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
	"fmt"
	"path/filepath"

	"github.com/sendgrid/sendgrid-go"
	"github.com/spf13/cobra"
)

// applyCmd represents the apply command
var applyCmd = &cobra.Command{
	Use:   "apply",
	Short: "Apply desired template file to Sendgrid ",
	Long: `Given structure of a Sendgrid transactional template:

(Folder) Name of email 
	(File) template.json
	(File) content.html
	(File) content.txt

Apply changes to Sendgrid application identified by its API_KEY stored on
enviroment variable SENDGRID_API_KEY pointing to template file.
As example:

sendsync apply -f templates/cool_email/template.json
`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("apply called")
		file, _ := cmd.Flags().GetString("file")
		fmt.Println("value of the flag name :" + file)

		dir := filepath.Dir(file)
		fileTemplate := getTemplateFromFile(file)
		activeVersionFile := findActiveVersion(fileTemplate)
		if activeVersionFile == nil {
			panic("no active version found on file, aborting")
		}

		html := readFile(filepath.Join(dir, "content.html"))
		plain := readFile(filepath.Join(dir, "content.txt"))
		// vamos a sendgrid

		targetTemplate := getTemplateByName(fileTemplate.Name)
		//fmt.Println(targetTemplate)
		if targetTemplate == nil {
			fmt.Println("No template found, creating...")
			request := sendgrid.GetRequest(apiKey, "/v3/templates", host)
			request.Method = "POST"

			templatePayload := template{
				Name:       fileTemplate.Name,
				Generation: fileTemplate.Generation,
			}

			templatePayloadJson, err := json.Marshal(templatePayload)
			if err != nil {
				panic(err)
			}

			request.Body = templatePayloadJson
			response, err := sendgrid.API(request)
			if err != nil {
				panic(err)
			}

			err = json.Unmarshal([]byte(response.Body), &targetTemplate)
			if err != nil {
				panic(err)
			}
		}

		activeVersion := findActiveVersion(*targetTemplate)
		fmt.Println(activeVersion)
		if activeVersion == nil {
			fmt.Println("no active version found, creating...")

			activeVersion = &version{
				Active:     1,
				Name:       activeVersionFile.Name,
				TemplateId: targetTemplate.Id,
				Subject:    activeVersionFile.Subject,
			}

			templatePayloadJson, err := json.Marshal(activeVersion)
			if err != nil {
				panic(err)
			}

			requestUri := fmt.Sprintf("/v3/templates/%s/versions", targetTemplate.Id)
			request := sendgrid.GetRequest(apiKey, requestUri, host)
			request.Body = templatePayloadJson
			request.Method = "POST"

			response, err := sendgrid.API(request)
			if err != nil {
				panic(err)
			}

			err = json.Unmarshal([]byte(response.Body), &activeVersion)
			if err != nil {
				panic(err)
			}
		}

		activeVersion.HtmlContent = html
		activeVersion.PlainContent = plain

		requestUri := fmt.Sprintf("/v3/templates/%s/versions/%s", targetTemplate.Id, activeVersion.Id)
		request := sendgrid.GetRequest(apiKey, requestUri, host)
		request.Method = "PATCH"

		versionJson, err := json.Marshal(activeVersion)
		if err != nil {
			fmt.Println(err)
		}

		request.Body = versionJson
		response, err := sendgrid.API(request)
		if err != nil {
			panic(err)
		} else {
			fmt.Println(response.StatusCode)
			fmt.Println(response.Body)
			fmt.Println(response.Headers)
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.PersistentFlags().StringP("file", "f", "", "Template manifest to apply")
	applyCmd.MarkPersistentFlagRequired("file")
}
