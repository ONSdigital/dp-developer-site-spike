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

type Layout struct {
	Title string
	Nav   []NavItem
}

type APITemplateData struct {
	Layout
	API
}

type PathTemplateData struct {
	Layout
	Path
}

type NavItem struct {
	Title string
	URI   string
}

type API struct {
	Title       string
	Version     string
	Description string
	Paths       []Path
	URL         string
}

type Path struct {
	Title   string
	Methods []Method
	URL     string
}

type methodType string

const (
	Get     methodType = "GET"
	Post    methodType = "POST"
	Put     methodType = "PUT"
	Delete  methodType = "DELETE"
	Head    methodType = "HEAD"
	Options methodType = "OPTIONS"
	Patch   methodType = "PATCH"
)

type Method struct {
	Type            methodType
	Description     string
	Parameters      []spec.ParamProps
	DefaultResponse *spec.Response
	Responses       []Response
}

type Response struct {
	*spec.Response
	StatusCode int
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

	var APIs []API

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

			spec.ExpandSpec(&APISpec, &spec.ExpandOptions{})

			APIData := API{
				Title:       APISpec.Info.Title,
				Version:     APISpec.Info.Version,
				Description: APISpec.Info.Description,
				URL:         makeSpecRootDirectory(sourceSpec.ID),
			}

			for key, value := range APISpec.Paths.Paths {
				APIData.Paths = append(APIData.Paths, Path{
					Title:   key,
					URL:     makeSpecPathDirectories(sourceSpec.ID, key),
					Methods: getMethodsData(value),
				})
			}

			APIs = append(APIs, APIData)
		}()
	}

	buildSpecAssets(renderer, APIs)
}

func convertMarkdownFileToHTML(path string) []byte {
	bytes, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Println(err)
		return nil
	}

	return bytes
}

func getMethodsData(pathItem spec.PathItem) []Method {
	var methods []Method
	if pathItem.Get != nil {
		methods = append(methods, Method{
			Type:            Get,
			Description:     pathItem.Get.Description,
			Parameters:      getParametersData(pathItem.Get.Parameters),
			DefaultResponse: pathItem.Get.Responses.Default,
			Responses:       getResponsesData(pathItem.Get.Responses.ResponsesProps.StatusCodeResponses),
		})
	}
	if pathItem.Post != nil {
		methods = append(methods, Method{
			Type:            Post,
			Description:     pathItem.Post.Description,
			Parameters:      getParametersData(pathItem.Post.Parameters),
			DefaultResponse: pathItem.Post.Responses.Default,
			Responses:       getResponsesData(pathItem.Post.Responses.ResponsesProps.StatusCodeResponses),
		})
	}
	if pathItem.Put != nil {
		methods = append(methods, Method{
			Type:            Put,
			Description:     pathItem.Put.Description,
			Parameters:      getParametersData(pathItem.Put.Parameters),
			DefaultResponse: pathItem.Put.Responses.Default,
			Responses:       getResponsesData(pathItem.Put.Responses.ResponsesProps.StatusCodeResponses),
		})
	}
	if pathItem.Delete != nil {
		methods = append(methods, Method{
			Type:            Delete,
			Description:     pathItem.Delete.Description,
			Parameters:      getParametersData(pathItem.Delete.Parameters),
			DefaultResponse: pathItem.Delete.Responses.Default,
			Responses:       getResponsesData(pathItem.Delete.Responses.ResponsesProps.StatusCodeResponses),
		})
	}
	if pathItem.Head != nil {
		methods = append(methods, Method{
			Type:            Head,
			Description:     pathItem.Head.Description,
			Parameters:      getParametersData(pathItem.Head.Parameters),
			DefaultResponse: pathItem.Head.Responses.Default,
			Responses:       getResponsesData(pathItem.Head.Responses.ResponsesProps.StatusCodeResponses),
		})
	}
	if pathItem.Options != nil {
		methods = append(methods, Method{
			Type:            Options,
			Description:     pathItem.Options.Description,
			Parameters:      getParametersData(pathItem.Options.Parameters),
			DefaultResponse: pathItem.Options.Responses.Default,
			Responses:       getResponsesData(pathItem.Options.Responses.ResponsesProps.StatusCodeResponses),
		})
	}
	if pathItem.Patch != nil {
		methods = append(methods, Method{
			Type:            Patch,
			Description:     pathItem.Patch.Description,
			Parameters:      getParametersData(pathItem.Patch.Parameters),
			DefaultResponse: pathItem.Patch.Responses.Default,
			Responses:       getResponsesData(pathItem.Patch.Responses.ResponsesProps.StatusCodeResponses),
		})
	}

	return methods
}

func getParametersData(parameters []spec.Parameter) []spec.ParamProps {
	var paramsProps []spec.ParamProps

	for _, parameter := range parameters {
		paramsProps = append(paramsProps, parameter.ParamProps)
	}

	return paramsProps
}

func getResponsesData(responses map[int]spec.Response) []Response {
	var responsesProps []Response

	for key, response := range responses {
		statusCode := key
		responsesProps = append(responsesProps, Response{
			&response,
			statusCode,
		})
	}

	return responsesProps
}

func makeSpecRootDirectory(ID string) string {
	directoryName := "assets/specs/" + ID
	os.MkdirAll(directoryName, os.FileMode(int(0777)))
	fmt.Printf("Creating directory: %q\n", directoryName)
	return strings.TrimPrefix(directoryName, "assets")
}

func makeSpecPathDirectories(ID string, path string) string {
	pathWithoutSlashes := strings.Replace(strings.TrimPrefix(path, "/"), "/", "-", -1)
	directoryName := "assets/specs/" + ID + "/" + pathWithoutSlashes
	os.MkdirAll(directoryName, os.FileMode(int(0777)))
	fmt.Printf("Creating path directory: %q\n", directoryName)
	return strings.TrimPrefix(directoryName, "assets")
}

func buildSpecAssets(renderer *render.Render, APIs []API) {
	for _, APIdata := range APIs {
		func() {
			file, err := os.Create("assets/" + APIdata.URL + "/index.html")
			if err != nil {
				log.Println(err)
				return
			}

			defer file.Close()

			err = renderer.HTML(file, 0, "api", APITemplateData{
				API:    APIdata,
				Layout: Layout{},
			})
			if err != nil {
				log.Println(err)
				return
			}
		}()

		buildSpecPathsAssets(renderer, APIdata.Paths)
	}
}

func buildSpecPathsAssets(renderer *render.Render, Paths []Path) {
	for _, pathData := range Paths {
		func() {
			file, err := os.Create("assets/" + pathData.URL + "/index.html")
			if err != nil {
				log.Println(err)
				return
			}

			defer file.Close()

			err = renderer.HTML(file, 0, "path", PathTemplateData{
				Path:   pathData,
				Layout: Layout{},
			})
			if err != nil {
				log.Println(err)
				return
			}
		}()
	}
}
