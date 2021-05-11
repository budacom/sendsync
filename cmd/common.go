package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"os"

	"github.com/sendgrid/sendgrid-go"
	log "github.com/sirupsen/logrus"
)

var apiKey = os.Getenv("SENDGRID_API_KEY")
var host = "https://api.sendgrid.com"

type Template struct {
	Generation string
	Name       string
	Id         string
	Versions   []Version
}

type Version struct {
	Id                   string
	TemplateId           string `json:"template_id"`
	Active               int
	Editor               string
	Name                 string
	ThumbnailUrl         string `json:"thumbnail_url"`
	HtmlContent          string `json:"html_content,omitempty"`
	PlainContent         string `json:"plain_content,omitempty"`
	Subject              string
	GeneratePlainContent bool `json:"generate_plain_content"`
}

type templates struct {
	Templates []Template
}

func writeFile(path string, content string) {
	err := ioutil.WriteFile(path, []byte(content), 0644)
	if err != nil {
		log.Println(err)
	}
}

func (template Template) fetchAndUpdateTemplate() {
	requestPath := fmt.Sprintf("/v3/templates/%s", template.Id)
	request := sendgrid.GetRequest(apiKey, requestPath, host)
	request.Method = "GET"
	response, err := sendgrid.API(request)
	if err != nil {
		log.Println(err)
	}
	template.UpdateTemplateFromJson(response.Body)
}

func (template *Template) UpdateTemplateFromJson(body string) {
	err := json.Unmarshal([]byte(body), &template)
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

func readFile(file string) string {
	fileContent, err := ioutil.ReadFile(file)
	if err != nil {
		panic(err)
	}
	return string(fileContent)
}

func getTemplateFromFile(file string) Template {
	var templateFromFile Template
	templateJson := readFile(file)
	templateFromFile.UpdateTemplateFromJson(templateJson)
	return templateFromFile
}

func getTemplateByName(name string) *Template {
	templates := getTemplates()
	var template *Template
	for _, template := range templates.Templates {
		if template.Name == name {
			return &template
		}
	}

	return template
}

func (template Template) FindActiveVersion() *Version {
	var version *Version
	for _, version := range template.Versions {
		if version.Active == 1 {
			return &version
		}
	}
	return version
}
