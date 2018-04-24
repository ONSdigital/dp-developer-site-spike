package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/go-openapi/spec"
	"github.com/unrolled/render"
)

type SiteStructure struct {
	Title string
	APIs  []API
	Nav   []NavItem
	Sitemap
}

type NavItem struct {
	Title string
	URI   string
}

type API struct {
	spec *spec.Swagger
}

type Sitemap struct {
}

func main() {
	specURLs := []string{
		"http://localhost:9900/dp-dataset-api/swagger.json",
		"http://localhost:9900/dp-import-api/swagger.json",
	}

	renderer := render.New(render.Options{
		Layout:     "layout",
		IndentJSON: true,
	})

	for index, specURL := range specURLs {
		func() {
			fmt.Printf("Getting swagger spec: %s\n", specURL)
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

			// TODO we should be creating the directory structure and storing as `index.html` in each one of those.
			file, err := os.Create("assets/" + strconv.Itoa(index) + ".html")
			if err != nil {
				log.Println(err)
				return
			}

			defer file.Close()

			renderer.HTML(file, 0, "api", APISpec)
		}()
	}
}
