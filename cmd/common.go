package cmd

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"

	"github.com/sendgrid/sendgrid-go"
)

var apiKey = os.Getenv("SENDGRID_API_KEY")
var host = "https://api.sendgrid.com"

/*type storedTemplate struct {
	Generation string
	Name       string
	Versions   []storedVersion
}

type storedVersion struct {
	Id                   string `json:"-"`
	TemplateId           string `json:"-"`
	Active               int
	Editor               string
	Name                 string
	ThumbnailUrl         string `json:"thumbnail_url"`
	HtmlContent          string `json:"html_content,omitempty"`
	PlainContent         string `json:"plain_content,omitempty"`
	Subject              string
	GeneratePlainContent bool `json:"generate_plain_content"`
}*/

type template struct {
	Generation string
	Name       string
	Id         string
	Versions   []version
}

// func (t template) isEmpty() bool {
// 	return t.Id == ""
// }

type version struct {
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
	Templates []template
}

/*func templateToStoredTemplate(template *template) *storedTemplate {
	var versions []storedVersion
	for _, version := range template.Versions {
		versions = append(versions, *versionToStoredVersion(version))
	}

	storedTemplate := &storedTemplate{
		Generation: template.Generation,
		Name:       template.Name,
		Versions:   versions,
	}
	return storedTemplate
}

func versionToStoredVersion(version *version) *storedVersion {
		Active:               version.Active,
		Editor:               version.Editor,
		Name:                 version.Name,
		ThumbnailUrl:         version.ThumbnailUrl,
		HtmlContent:          version.HtmlContent,
		PlainContent:         version.PlainContent,
		Subject:              version.Subject,
		GeneratePlainContent: version.GeneratePlainContent,
	}
	return storedVersion
}*/

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

func getTemplateByName(name string) *template {
	templates := getTemplates()
	var template *template
	for _, template := range templates.Templates {
		if template.Name == name {
			return &template
		}
	}

	return template
}

func findActiveVersion(template template) *version {
	var version *version
	for _, version := range template.Versions {
		if version.Active == 1 {
			return &version
		}
	}
	return version
}
