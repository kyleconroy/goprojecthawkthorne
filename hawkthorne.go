package main

import (
	"encoding/base64"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"net/url"
	"os"
)

func renderTemplate(path string) func(http.ResponseWriter, *http.Request) {
	tmpl, err := template.ParseFiles("templates/base.html", "templates/"+path)

	if err != nil {
		log.Fatal(err)
	}

	return func(w http.ResponseWriter, r *http.Request) {
		tmpl.Execute(w, nil)
	}
}

func trackDownload(path string) func(http.ResponseWriter, *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		go func() {
			datum := map[string]interface{}{
				"event": "game.downloaded",
				"properties": map[string]interface{}{
					"token": os.Getenv("MIXPANEL_TOKEN"),
					"ip":    r.Header.Get("X-Forwarded-For"),
				},
			}

			b, err := json.Marshal(datum)

			if err != nil {
				return
			}

			v := url.Values{}
			v.Set("data", base64.StdEncoding.EncodeToString(b))
			_, _ = http.Get("http://api.mixpanel.com/track/?" + v.Encode())
		}()

		url := "https://s3.amazonaws.com/hawkthorne.journey.builds/" + path
		http.Redirect(w, r, url, http.StatusFound)
	}
}

func main() {
	http.HandleFunc("/", renderTemplate("index.html"))
	http.HandleFunc("/sprites", renderTemplate("sprites.html"))
	http.HandleFunc("/soundtrack/disc/1", renderTemplate("disc1.html"))
	http.HandleFunc("/soundtrack/disc/2", renderTemplate("disc2.html"))

	http.HandleFunc("/downloads/hawkthorne-win-x64.zip", trackDownload("hawkthorne-win-x64.zip"))
	http.HandleFunc("/downloads/hawkthorne-win-x86.zip", trackDownload("hawkthorne-win-x86.zip"))
	http.HandleFunc("/downloads/hawkthorne-osx.zip", trackDownload("hawkthrone-osx.zip"))
	http.HandleFunc("/downloads/hawkthorne.love", trackDownload("hawkthorne.love"))

	http.Handle("/sprites.html", http.RedirectHandler("/sprites", 301))
	http.Handle("/audio.html", http.RedirectHandler("/soundtrack/disc/1", 301))
	http.Handle("/soundtrack.html", http.RedirectHandler("/soundtrack/disc/1", 301))
	http.Handle("/soundtrack-disc2.html", http.RedirectHandler("/soundtrack/disc/2", 301))

	port := os.Getenv("PORT")

	if port == "" {
		port = "8080"
	}

	pfx := "/static/"

	h := http.StripPrefix(pfx, http.FileServer(http.Dir("static")))

	http.Handle(pfx, h)

	log.Fatal(http.ListenAndServe(":"+port, nil))
}
