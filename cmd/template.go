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

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var templateCmd = &cobra.Command{
	Use:   "template",
	Short: "Fetch templates from sendgrid",
	Long: `Fetch all templates from a Sendgrid Application identified by SENDGRID_API_KEY
	enviroment variable`,
	Run: func(cmd *cobra.Command, args []string) {
		if verbose {
			log.SetLevel(log.TraceLevel)
		}

		log.Info("Started retrieving templates")

		templates := getTemplates()
		makeDir("templates")
		for _, template := range templates.Templates {
			log.WithFields(log.Fields{
				"template": template,
			}).Debug("Processing template")

			dirPath := fmt.Sprintf("templates/%s", template.Name)
			makeDir(dirPath)
			templateJson, err := json.MarshalIndent(template, "", "  ")
			if err != nil {
				log.WithFields(log.Fields{
					"template": template,
				}).Warn(err)
			}
			writeFile(fmt.Sprintf("%s/template.json", dirPath), string(templateJson))
			template.fetchAndUpdateTemplate()
			activeVersion := template.FindActiveVersion()
			if activeVersion == nil {
				log.WithFields(log.Fields{
					"template": template,
				}).Warn("no active version found")
			} else {
				writeFile(fmt.Sprintf("%s/content.html", dirPath), activeVersion.HtmlContent)
				writeFile(fmt.Sprintf("%s/content.txt", dirPath), activeVersion.PlainContent)
			}
		}
		log.Info("Finished retrieving templates")
	},
}

func init() {
	getCmd.AddCommand(templateCmd)
}
