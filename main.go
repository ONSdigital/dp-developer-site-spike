package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-openapi/spec"
	"github.com/russross/blackfriday"
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

type sourceSpec struct {
	ID, URL string
}

func main() {
	sourceSpecs := []sourceSpec{
		{"dataset-api", "http://localhost:9900/dp-dataset-api/swagger.json"},
		{"import-api", "http://localhost:9900/dp-import-api/swagger.json"},
	}

	renderer := render.New(render.Options{
		Layout:     "layout",
		IndentJSON: true,
	})

	err := filepath.Walk("static", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("prevent panic by handling failure accessing a path %q: %v\n", "static", err)
			return err
		}
		if info.IsDir() {
			fmt.Printf("Creating directory: %q\n", path)
			os.MkdirAll("assets"+strings.TrimPrefix(path, "static"), info.Mode())
			return nil
		}

		markdownNameParts := strings.SplitAfter(info.Name(), ".")
		if markdownNameParts[1] == "md" || markdownNameParts[1] == "markdown" {
			convertedMarkdown := blackfriday.MarkdownBasic(convertMarkdownFileToHTML(path))
			fileDirectory := "assets" + strings.TrimPrefix(path, "static")
			newFilePath := strings.TrimSuffix(fileDirectory, info.Name())
			fmt.Println(newFilePath)
			err := ioutil.WriteFile(newFilePath+markdownNameParts[0]+"html", convertedMarkdown, 0644)
			if err != nil {
				fmt.Println(err)
				return nil
			}
		}

		return nil
	})

	if err != nil {
		fmt.Printf("error walking the path %q: %v\n", "static", err)
		return
	}

	for _, sourceSpec := range sourceSpecs {
		func() {
			fmt.Printf("Getting swagger spec: %s\n", sourceSpec.URL)
			req, err := http.Get(sourceSpec.URL)
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

			os.MkdirAll("assets/specs/"+sourceSpec.ID, os.FileMode(int(0777)))
			fmt.Printf("Creating directory: %q\n", "assets/specs/"+sourceSpec.ID)

			spec.ExpandSpec(&APISpec, &spec.ExpandOptions{})

			file, err := os.Create("assets/specs/" + sourceSpec.ID + "/index.html")
			if err != nil {
				log.Println(err)
				return
			}

			defer file.Close()

			renderer.HTML(file, 0, "api", APISpec)
		}()
	}
}

func convertMarkdownFileToHTML(path string) []byte {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return bytes
}
