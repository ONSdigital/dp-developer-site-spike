package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/alecthomas/template"
	"github.com/go-openapi/spec"
)

func main() {
	specURLs := []string{
		"http://localhost:9900/dp-dataset-api/swagger.json",
		"http://localhost:9900/dp-import-api/swagger.json",
	}

	for index, specURL := range specURLs {
		fmt.Printf("Getting swagger spec: %s", specURL)
		req, err := http.Get(specURL)
		if err != nil {
			log.Println(err)
			return
		}
		defer req.Body.Close()

		bodyBytes, err := ioutil.ReadAll(req.Body)
		if err != nil {
			log.Println(err)
			return
		}

		var APISpec spec.Swagger
		err = json.Unmarshal(bodyBytes, &APISpec)
		if err != nil {
			log.Println(err)
			return
		}

		spec.ExpandSpec(&APISpec, &spec.ExpandOptions{})

		tmpl, err := template.ParseFiles("templates/index.tmpl")
		if err != nil {
			log.Println(err)
			return
		}

		file, err := os.Create("assets/" + strconv.Itoa(index) + ".html")
		if err != nil {
			log.Println(err)
			return
		}

		defer file.Close()

		err = tmpl.Execute(file, APISpec)
		if err != nil {
			log.Print("execute: ", err)
			return
		}
	}
}
