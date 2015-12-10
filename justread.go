package main

import (
	"bytes"
	"fmt"
	"github.com/russross/blackfriday"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"regexp"
	"runtime"
	"strings"
	"time"
)

// https://play.golang.org/p/N8A5cXa3RM
func wordCountWithoutLinks(data_string string) int {
	// from https://regex101.com/r/tZ6yK9/3
	data := data_string
	re := regexp.MustCompile("\\[([^\\]]+)\\]\\(([^\\)]+)\\)")
	matches := re.FindAllString(data, -1)
	newdata := data
	for _, match := range matches {
		if match[0:2] == "[!" {
			match = match[2:]
		}
		newdata = strings.Replace(newdata, match, "", -1)
	}
	data = newdata
	matches = re.FindAllString(data, -1)
	newdata = data
	for _, match := range matches {
		if match[0:3] == "[![" {
			match = match[2:]
		}
		newdata = strings.Replace(newdata, match, "", -1)
	}
	return len(strings.Split(strings.TrimSpace(newdata), " "))
}

func timeTrack(start time.Time, name string) {
	elapsed := time.Since(start)
	log.Printf("%s took %s", name, elapsed)
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func downloadUrl(url string) string {
	defer timeTrack(time.Now(), "downloadUrl")
	response, err := http.Get(url)
	if err != nil {
		fmt.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			fmt.Printf("%s", err)
			os.Exit(1)
		}
		fmt.Printf("%s\n", string(contents))
		f, err := os.Create("test.html")
		check(err)
		defer f.Close()
		_, err = f.Write(contents)
		check(err)

	}
	return "test.html"
}

func parseURL(url string) string {
	defer timeTrack(time.Now(), "wget")
	var (
		cmdOut []byte
		err    error
	)
	cmdName := "pandoc"
	cmdArgs := []string{"--columns=70000", "-t", "markdown", "-s", downloadUrl(url)}
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running git rev-parse command: ", err)
		os.Exit(1)
	}
	sha := string(cmdOut)
	defer timeTrack(time.Now(), "parsing")
	result := strings.Split(sha, "\n")

	var w bytes.Buffer

	lastSentenceGood := false
	for i := range result {
		if i < 3 {
			if strings.Contains(result[i],"title:") {
				w.WriteString("# ")
				w.WriteString(strings.Split(strings.TrimSpace(result[i]), "title:")[1])
				w.WriteString("\n\n")
			}
		} else {
			strLen := len(strings.Split(strings.TrimSpace(result[i]), " "))
			strLen2 := wordCountWithoutLinks(result[i])

			fmt.Printf("%d) %d/%d: %s\n", i, strLen, strLen2, result[i])
			if strLen2 > 10 {
				w.WriteString(result[i])
				w.WriteString("\n\n")
				lastSentenceGood = true
			} else if lastSentenceGood == true && strLen2 > 3 {
				w.WriteString(result[i])
				w.WriteString("\n\n")
				lastSentenceGood = false
			} else if strLen > 1 {
				lastSentenceGood = false
			}
		}
	}

	fmt.Println("WEB RESULT:\n\n")
	output := blackfriday.MarkdownCommon(w.Bytes())
	return string(output)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	defer timeTrack(time.Now(), "indexHandler")
	url := r.URL.Query()["url"]
	if len(url) > 0 {
		result := parseURL(url[0])
		fmt.Fprintln(w, result)
	}
	fmt.Fprintln(w, "No url")
}

func startServer() {
	runtime.GOMAXPROCS(runtime.NumCPU() - 1) // one core for wrk

	http.HandleFunc("/", indexHandler)
	http.ListenAndServe(":4000", nil)
	fmt.Println("Fin Bench running on Port 4000")
}
func main() {
	startServer()
}
