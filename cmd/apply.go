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
	"log"
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
		file, _ := cmd.Flags().GetString("file")
		dir := filepath.Dir(file)
		fileTemplate := getTemplateFromFile(file)
		activeVersionFile := findActiveVersion(fileTemplate)

		if activeVersionFile == nil {
			log.Fatal("No active version found on file, aborting")
		}

		html := readFile(filepath.Join(dir, "content.html"))
		plain := readFile(filepath.Join(dir, "content.txt"))

		log.Printf("Retriving template %s from API", fileTemplate.Name)
		targetTemplate := getTemplateByName(fileTemplate.Name)

		if targetTemplate == nil {
			log.Printf("No template %s found, creating...", fileTemplate.Name)
			request := sendgrid.GetRequest(apiKey, "/v3/templates", host)
			request.Method = "POST"

			templatePayload := template{
				Name:       fileTemplate.Name,
				Generation: fileTemplate.Generation,
			}

			templatePayloadJson, err := json.Marshal(templatePayload)
			if err != nil {
				log.Fatal(err)
			}
			request.Body = templatePayloadJson
			response, err := sendgrid.API(request)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("Got %d status from API when creating template %s", response.StatusCode, templatePayload.Name)
			err = json.Unmarshal([]byte(response.Body), &targetTemplate)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Template %s created succesfully", targetTemplate.Name)
		}

		activeVersion := findActiveVersion(*targetTemplate)

		if activeVersion == nil {
			log.Printf("No active version found for template %s, creating...", targetTemplate.Name)

			activeVersion = &version{
				Active:     1,
				Name:       activeVersionFile.Name,
				TemplateId: targetTemplate.Id,
				Subject:    activeVersionFile.Subject,
			}

			templatePayloadJson, err := json.Marshal(activeVersion)
			if err != nil {
				log.Fatal(err)
			}

			requestUri := fmt.Sprintf("/v3/templates/%s/versions", targetTemplate.Id)
			request := sendgrid.GetRequest(apiKey, requestUri, host)
			request.Body = templatePayloadJson
			request.Method = "POST"

			response, err := sendgrid.API(request)
			if err != nil {
				log.Fatal(err)
			}

			log.Printf("Got %d status from API when creating version template %s", response.StatusCode, activeVersion.Name)
			err = json.Unmarshal([]byte(response.Body), &activeVersion)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Version %s for template %s created succesfully", activeVersion.Name, targetTemplate.Name)
		}

		activeVersion.HtmlContent = html
		activeVersion.PlainContent = plain

		requestUri := fmt.Sprintf("/v3/templates/%s/versions/%s", targetTemplate.Id, activeVersion.Id)
		request := sendgrid.GetRequest(apiKey, requestUri, host)
		request.Method = "PATCH"

		versionJson, err := json.Marshal(activeVersion)
		if err != nil {
			log.Fatal(err)
		}

		request.Body = versionJson

		log.Printf("Updating template %s with version %s", targetTemplate.Name, activeVersion.Name)
		response, err := sendgrid.API(request)
		if err != nil {
			log.Fatal(err)
		} else {
			log.Printf("Got %d status from API when updating template %s with version %s", response.StatusCode, targetTemplate.Name, activeVersion.Name)
		}
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.PersistentFlags().StringP("file", "f", "", "Template manifest to apply")
	applyCmd.MarkPersistentFlagRequired("file")
}
