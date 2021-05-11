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
	log "github.com/sirupsen/logrus"
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
		if verbose {
			log.SetLevel(log.TraceLevel)
		}

		file, _ := cmd.Flags().GetString("file")
		dir := filepath.Dir(file)
		fileTemplate := getTemplateFromFile(file)
		activeVersionFile := findActiveVersion(fileTemplate)

		log.WithFields(log.Fields{
			"filename": file,
		}).Info("Started applying changes in template")

		if activeVersionFile == nil {
			log.WithFields(log.Fields{
				"template": fileTemplate,
			}).Fatal("No active version found on file, aborting")
		}

		html := readFile(filepath.Join(dir, "content.html"))
		plain := readFile(filepath.Join(dir, "content.txt"))

		log.WithFields(log.Fields{
			"template": fileTemplate,
		}).Debug("Retriving template from API")
		targetTemplate := getTemplateByName(fileTemplate.Name)

		if targetTemplate == nil {
			log.WithFields(log.Fields{
				"template": fileTemplate,
			}).Debug("No remote template found, creating...")
			request := sendgrid.GetRequest(apiKey, "/v3/templates", host)
			request.Method = "POST"

			templatePayload := template{
				Name:       fileTemplate.Name,
				Generation: fileTemplate.Generation,
			}

			templatePayloadJson, err := json.Marshal(templatePayload)
			if err != nil {
				log.WithFields(log.Fields{
					"template": templatePayload,
				}).Fatal(err)
			}
			request.Body = templatePayloadJson

			log.WithFields(log.Fields{
				"request_uri": request.BaseURL,
				"template":    templatePayload,
			}).Debug("Calling API to create template")

			response, err := sendgrid.API(request)
			if err != nil {
				log.WithFields(log.Fields{
					"request": request,
				}).Fatal(err)
			}

			err = json.Unmarshal([]byte(response.Body), &targetTemplate)
			if err != nil {
				log.WithFields(log.Fields{
					"body":     response.Body,
					"template": targetTemplate,
				}).Fatal(err)
			}
			log.WithFields(log.Fields{
				"response": response,
				"template": targetTemplate,
			}).Debug("Template created succesfully")
		}

		activeVersion := findActiveVersion(*targetTemplate)

		if activeVersion == nil {
			log.WithFields(log.Fields{
				"template": targetTemplate,
			}).Debug("No active version found for template, creating...")

			activeVersion = &version{
				Active:     1,
				Name:       activeVersionFile.Name,
				TemplateId: targetTemplate.Id,
				Subject:    activeVersionFile.Subject,
			}

			templatePayloadJson, err := json.Marshal(activeVersion)
			if err != nil {
				log.WithFields(log.Fields{
					"version": activeVersion,
				}).Fatal(err)
			}

			requestUri := fmt.Sprintf("/v3/templates/%s/versions", targetTemplate.Id)
			request := sendgrid.GetRequest(apiKey, requestUri, host)
			request.Body = templatePayloadJson
			request.Method = "POST"

			log.WithFields(log.Fields{
				"request_uri": request.BaseURL,
				"template":    activeVersion,
			}).Debug("Calling API to create version")

			response, err := sendgrid.API(request)
			if err != nil {
				log.WithFields(log.Fields{
					"request": request,
				}).Fatal(err)
			}
			err = json.Unmarshal([]byte(response.Body), &activeVersion)
			if err != nil {
				log.WithFields(log.Fields{
					"version": activeVersion,
				}).Fatal(err)
			}
			log.WithFields(log.Fields{
				"response": response,
				"version":  activeVersion,
			}).Debug("Version created successfully")
		}

		activeVersion.HtmlContent = html
		activeVersion.PlainContent = plain

		requestUri := fmt.Sprintf("/v3/templates/%s/versions/%s", targetTemplate.Id, activeVersion.Id)
		request := sendgrid.GetRequest(apiKey, requestUri, host)
		request.Method = "PATCH"

		versionJson, err := json.Marshal(activeVersion)
		if err != nil {
			log.WithFields(log.Fields{
				"version": activeVersion,
			}).Fatal(err)
		}

		request.Body = versionJson

		log.WithFields(log.Fields{
			"request_uri": request.BaseURL,
			"version":     activeVersion,
		}).Debug("Calling API to update template version")
		response, err := sendgrid.API(request)
		if err != nil {
			log.WithFields(log.Fields{
				"request": request,
			}).Fatal(err)
		} else {
			log.WithFields(log.Fields{
				"response": response,
				"version":  activeVersion,
			}).Debug("Version updated successfully")
		}

		log.WithFields(log.Fields{
			"template_name": targetTemplate.Name,
			"version_name":  activeVersion.Name,
		}).Info("Finished applying changes in template")
	},
}

func init() {
	rootCmd.AddCommand(applyCmd)
	applyCmd.PersistentFlags().StringP("file", "f", "", "Template manifest to apply")
	applyCmd.MarkPersistentFlagRequired("file")
}
