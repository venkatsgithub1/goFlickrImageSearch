package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

// MyMux struct handles routes.
type MyMux struct{}

// RandomImage struct takes flickr image url retrieved.
type RandomImage struct {
	TagData   string
	ImageData Response
}

// Body contains response retrieved from server.
type Response struct {
	Photos struct {
		Page    int
		Pages   int
		PerPage int
		Total   string
		Photo   []struct {
			ID       string
			Owner    string
			Secret   string
			Server   string
			Title    string
			Farm     int
			Ispublic int
			Isfamily int
			Isfriend int
			Url_M    string
			Height_M string
			Width_M  string
		}
	}
	Stat string
}

var randomImg = RandomImage{}

// ServeHTTP function handles routes.
func (p *MyMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	//fmt.Println("path:", r.URL.Path)
	if r.URL.Path == "/" {
		getTagData(w, r)
		return
	}
	if r.URL.Path == "/image" {
		getPublicPhotos(w, r)
		return
	}

	http.NotFound(w, r)
}

func renderTemplate(w http.ResponseWriter, data interface{}, htmlPage string) {
	tmpl, err := template.ParseFiles("./templates/" + htmlPage + ".html")

	checkErr(err)

	// randomImg's url will be shown on page.
	//fmt.Println("executing template")
	switch t := data.(type) {
	case RandomImage:
		//fmt.Println(t)
		tmpl.Execute(w, t)
		return
	}
	tmpl.Execute(w, data)
}

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func getTagData(w http.ResponseWriter, r *http.Request) {
	// calling parseForm to get form data.
	renderTemplate(w, "", "mainPage")
}

func getPublicPhotos(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	tagData := r.Form["tagData"]
	tagDataStr := strings.Join(tagData, ",")
	//fmt.Println("before if: tagData", tagData)
	// get tag data.
	if randomImg.TagData == "" || tagDataStr != "" && randomImg.TagData != tagDataStr {
		randomImg.TagData = tagDataStr
		//fmt.Println("inside if")
		//fmt.Println(randomImg, "sending this data")
		searchImages(&randomImg)
	}
	renderTemplate(w, randomImg, "showImage")
}

func searchImages(randomImg *RandomImage) {
	tagDataForURL := strings.Replace(randomImg.TagData, " ", "%20", -1)
	url := fmt.Sprintf("https://api.flickr.com/services/rest/?method=flickr.photos.search&api_key=8abd252e106c4c416522b37a1e2d722f&tags=%s&privacy_filter=1&format=json&nojsoncallback=1&extras=url_m", tagDataForURL)

	req, err := http.NewRequest("GET", url, nil)

	checkErr(err)

	client := &http.Client{}

	resp, err := client.Do(req)

	body := resp.Body

	defer body.Close()

	flickrBytes, err := ioutil.ReadAll(body)

	checkErr(err)

	jsonResp := &Response{}

	json.Unmarshal(flickrBytes, jsonResp)

	//fmt.Println("jsonResp after: ", jsonResp)

	randomImg.ImageData = *jsonResp
}

func main() {
	mux := &MyMux{}
	fmt.Println("port from main.go: " + os.Getenv("PORT"))
	http.ListenAndServe(":"+os.Getenv("PORT"), mux)
}
