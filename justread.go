package main

import (
	"bufio"
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
	re := regexp.MustCompile("\\[(.*?)\\]\\((.+?)\\)")
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

func processString(data string) string {
	re := regexp.MustCompile("\\{([^}]+)\\}")
	newdata := data
	for {

		matches := re.FindAllString(newdata, -1)
		if len(matches) == 0 {
			break
		}

		for _, match := range matches {
			newdata = strings.Replace(newdata, match, "", -1)
		}
	}
	return newdata
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
	response, err := http.Get(url)
	//var re = regexp.MustCompile(`</?\w+((\s+\w+(\s*=\s*(?:".*?"|'.*?'|[^'">\s]+))?)+\s*|\s*)/?>`)
	if err != nil {
		log.Printf("%s", err)
		os.Exit(1)
	} else {
		defer response.Body.Close()
		contents, err := ioutil.ReadAll(response.Body)
		if err != nil {
			log.Printf("%s", err)
			os.Exit(1)
		}
		f, err := os.Create("test.html")
		check(err)
		defer f.Close()
		content_string := string(contents[:])
		result := strings.Split(content_string, "\n")
		var buffer bytes.Buffer
		for i := range result {
			newString := strings.Replace(result[i], "<", "\n<", -1)
			buffer.WriteString(strings.Replace(newString, ">", ">\n", -1))
		}

		result = strings.Split(buffer.String(), "\n")
		w := bufio.NewWriter(f)
		for i := range result {
			if i > 0 {
				j := i - 1
				if len(result[j]) > 500 ||
					strings.Contains(result[j], "div>") ||
					strings.Contains(result[j], "<div") ||
					strings.Contains(result[j], "</div") ||
					strings.Contains(result[j], "<meta") ||
					strings.Contains(result[j], "<form") ||
					strings.Contains(result[j], "</form") ||
					strings.Contains(result[j], "<img") ||
					strings.Contains(result[j], "<img") ||
					strings.Contains(result[j], "</img") ||
          strings.Contains(result[j], "<input") ||
					(strings.TrimSpace(strings.Replace(result[i], "/", "", -1)) == strings.TrimSpace(result[j])) ||
					len(result[j]) < 3 {

				} else {
					w.WriteString(strings.TrimSpace(result[j]))
					w.WriteString("\n")

				}

			}
		}

	}
	return "test.html"
}

func parseURL(url string) string {
	start := time.Now()
	file := downloadUrl(url)
	elapsed := time.Since(start)
	log.Printf("downloading took %s", elapsed)

	start = time.Now()
	var (
		cmdOut []byte
		err    error
	)
	cmdName := "pandoc"
	cmdArgs := []string{"--columns=70000", "-t", "markdown", "-s", file}
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		log.Println("There was an error running: ", err)
		os.Exit(1)
	}
	sha := string(cmdOut)
	result := strings.Split(sha, "\n")

	elapsed = time.Since(start)
	log.Printf("pandoc took %s", elapsed)

	var w bytes.Buffer

	lastSentenceGood := false
	for i := range result {
		result[i] = strings.TrimSpace(processString(result[i]))
		if i < 3 {
			if strings.Contains(result[i], "title:") {
				w.WriteString("# ")
				w.WriteString(strings.Split(result[i], "title:")[1])
				w.WriteString("\n\n")
			}
		} else {
			strLen := len(strings.Split(result[i], " "))
			strLen2 := wordCountWithoutLinks(result[i])

			//fmt.Printf("%d) %d/%d: %s\n", i, strLen, strLen2, result[i])
			if strLen2 > 20 && !(
          strings.Contains(result[i], "Share on Twitter") ||
          strings.Contains(result[i], "Sign in to") ||
          strings.Contains(result[i], "Sign in or") ||
          strings.Contains(result[i], "Comments") ||
          strings.Contains(result[i], "Play Videos") ||
          strings.Contains(result[i], "Your comment:") ||
          strings.Contains(result[i], "You’ll receive free")) {
				w.WriteString(result[i])
				w.WriteString("\n\n")
				lastSentenceGood = true
			} else if lastSentenceGood == true && strLen2 > 5 {
				w.WriteString(result[i])
				w.WriteString("\n\n")
				lastSentenceGood = false
			} else if strLen > 1 {
				lastSentenceGood = false
			}
		}
	}

	output := blackfriday.MarkdownCommon(w.Bytes())

	front := `<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="UTF-8">
         <link rel="shortcut icon" sizes="16x16 24x24 32x32 48x48 64x64" href="/static/img/favicon.ico" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0, user-scalable=no">
    <title>Go Just Read</title>
    <style type="text/css">
      body {
        margin: 0;
        padding: 0.4em 1em 6em;
        background: #fff;
      }
      .yue {
        max-width: 650px;
        margin: 0 auto;
      }
/**
 * yue.css
 *
 * yue.css is designed for readable content.
 *
 * Copyright (c) 2013 - 2014 by Hsiaoming Yang.
 */

.yue {
  font: 400 18px/1.62 "Georgia", "Xin Gothic", "Hiragino Sans GB", "Droid Sans Fallback", "Microsoft YaHei", sans-serif;
  color: #1C1C1B;  /* #444443;*/
}

.windows .yue {
  font-size: 16px;
  font-family: "Georgia", "SimSun", sans-serif;
}

.yue ::-moz-selection {
  background-color: rgba(0,0,0,0.2);
}

.yue ::selection {
  background-color: rgba(0,0,0,0.2);
}

.yue h1,
.yue h2,
.yue h3,
.yue h4,
.yue h5,
.yue h6 {
  font-family: "Georgia", "Xin Gothic", "Hiragino Sans GB", "Droid Sans Fallback", "Microsoft YaHei", "SimSun", sans-serif;
  color: #222223;
}

.yue h1 {
  font-size: 1.8em;
  margin: 0.67em 0;
}

.yue > h1 {
  margin-top: 0;
  font-size: 2em;
}

.yue h2 {
  font-size: 1.5em;
  margin: 0.83em 0;
}

.yue h3 {
  font-size: 1.17em;
  margin: 1em 0;
}

.yue h4,
.yue h5,
.yue h6 {
  font-size: 1em;
  margin: 1.6em 0 1em 0;
}

.yue h6 {
  font-weight: 500;
}

.yue p {
  margin-top: 0;
  margin-bottom: 1.46em;
}

.yue a {
  color: #111;
  word-wrap: break-word;
  -moz-text-decoration-color: rgba(0, 0, 0, 0.4);
  text-decoration-color: rgba(0, 0, 0, 0.4);
}

.yue a:hover {
  color: #555;
  -moz-text-decoration-color: rgba(0, 0, 0, 0.6);
  text-decoration-color: rgba(0, 0, 0, 0.6);
}

.yue h1 a,
.yue h2 a,
.yue h3 a {
  text-decoration: none;
}

.yue strong,
.yue b {
  font-weight: 700;
  color: #222223;
}

.yue em,
.yue i {
  font-style: italic;
  color: #222223;
}

.yue img {
  max-width: 100%;
  height: auto;
  margin: 0.2em 0;
}

.yue a img {
  /* Remove border on IE */
  border: none;
}

.yue figure {
  position: relative;
  clear: both;
  outline: 0;
  margin: 10px 0 30px;
  padding: 0;
  min-height: 100px;
}

.yue figure img {
  display: block;
  max-width: 100%;
  margin: auto auto 4px;
  box-sizing: border-box;
}

.yue figure figcaption {
  position: relative;
  width: 100%;
  text-align: center;
  left: 0;
  margin-top: 10px;
  font-weight: 400;
  font-size: 14px;
  color: #666665;
}

.yue figure figcaption a {
  text-decoration: none;
  color: #666665;
}

.yue hr {
  display: block;
  width: 14%;
  margin: 40px auto 34px;
  border: 0 none;
  border-top: 3px solid #dededc;
}

.yue blockquote {
  margin: 0 0 1.64em 0;
  border-left: 3px solid #dadada;
  padding-left: 12px;
  color: #666664;
}

.yue blockquote a {
  color: #666664;
}

.yue ul,
.yue ol {
  margin: 0 0 24px 6px;
  padding-left: 16px;
}

.yue ul {
  list-style-type: square;
}

.yue ol {
  list-style-type: decimal;
}

.yue li {
  margin-bottom: 0.2em;
}

.yue li ul,
.yue li ol {
  margin-top: 0;
  margin-bottom: 0;
  margin-left: 14px;
}

.yue li ul {
  list-style-type: disc;
}

.yue li ul ul {
  list-style-type: circle;
}

.yue li p {
  margin: 0.4em 0 0.6em;
}

.yue .unstyled {
  list-style-type: none;
  margin: 0;
  padding: 0;
}

.yue code,
.yue tt {
  color: #808080;
  font-size: 0.96em;
  background-color: #f9f9f7;
  padding: 1px 2px;
  border: 1px solid #dadada;
  border-radius: 3px;
  font-family: Menlo, Monaco, Consolas, "Courier New", monospace;
  word-wrap: break-word;
}

.yue pre {
  margin: 1.64em 0;
  padding: 7px;
  border: none;
  border-left: 3px solid #dadada;
  padding-left: 10px;
  overflow: auto;
  line-height: 1.5;
  font-size: 0.96em;
  font-family: Menlo, Monaco, Consolas, "Courier New", monospace;
  color: #4c4c4c;
  background-color: #f9f9f7;
}

.yue pre code,
.yue pre tt {
  color: #4c4c4c;
  border: none;
  background: none;
  padding: 0;
}

.yue table {
  width: 100%;
  max-width: 100%;
  border-collapse: collapse;
  border-spacing: 0;
  margin-bottom: 1.5em;
  font-size: 0.96em;
  box-sizing: border-box;
}

.yue th,
.yue td {
  text-align: left;
  padding: 4px 8px 4px 10px;
  border: 1px solid #dadada;
}

.yue td {
  vertical-align: top;
}

.yue tr:nth-child(even) {
  background-color: #efefee;
}

.yue iframe {
  display: block;
  max-width: 100%;
  margin-bottom: 30px;
}

.yue figure iframe {
  margin: auto;
}

.yue table pre {
  margin: 0;
  padding: 0;
  border: none;
  background: none;
}

@media (min-width: 1100px) {
  .yue blockquote {
    margin-left: -24px;
    padding-left: 20px;
    border-width: 4px;
  }

  .yue blockquote blockquote {
    margin-left: 0;
  }
}


    </style>
    
  </head>
  <body>
<div class="yue">
<p align="center"><a href="/">&larr; Another</a> &nbsp; &nbsp;<small><a href="ORIGINAL_URL">Original.</a></small></p>
<hr>
`

	back := `

<p align="center"><a href="/">&larr; Another</a></p>

</div>

</body>
</html>`
	s := []string{front, string(output), back}
	s2 := strings.Replace(string(strings.Join(s, " ")), "ORIGINAL_URL", url, -1)
	return s2
}

func indexHandler(w http.ResponseWriter, r *http.Request) {

	indexPage := `
<!DOCTYPE html>
<html>
  <head>
    <meta charset=utf-8 />
    <title>Just Read</title>
         <link rel="shortcut icon" sizes="16x16 24x24 32x32 48x48 64x64" href="/static/img/favicon.ico" />
    <meta name='viewport' content='initial-scale=1,maximum-scale=1,user-scalable=no' />
    
    <!-- Bootstrap and JQuery JS -->
      <script src="https://code.jquery.com/jquery-2.1.4.min.js" ></script>
    <script src="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/js/bootstrap.min.js" integrity="sha256-KXn5puMvxCw+dAYznun+drMdG1IFl3agK0p/pqT9KAo= sha512-2e8qq0ETcfWRI4HJBzQiA3UoyFk6tbNyG+qSaIBZLyW9Xf3sWZHN/lxe9fTh1U45DpPf07yj94KsUHHWe4Yk1A==" crossorigin="anonymous"></script>

    <!-- Bootstrap Core CSS -->
    <link href="https://maxcdn.bootstrapcdn.com/bootstrap/3.3.6/css/bootstrap.min.css" rel="stylesheet" integrity="sha256-7s5uDGW3AHqw6xtJmNNtr+OBRJUlgkNJEo78P4b0yRw= sha512-nNo+yCHEyn0smMxSswnf/OnX6/KwJuZTlNZBjauKhTK0c+zT+q5JOCx0UFhXQ6rJR9jg6Es8gPuD2uZcYDLqSw==" crossorigin="anonymous">



  <style>
/*!
 * Start Bootstrap - 2 Col Portfolio HTML Template (http://startbootstrap.com)
 * Code licensed under the Apache License v2.0.
 * For details, see http://www.apache.org/licenses/LICENSE-2.0.
 */

body {
    padding-top: 70px; /* Required padding for .navbar-fixed-top. Remove if using .navbar-static-top. Change if height of navigation changes. */
}

.portfolio-item {
    margin-bottom: 25px;
}

footer {
    margin: 50px 0;
}


header, main { padding: 0 20px; }

/*** wrapper div for both header and main ***/
.wrapper { margin-top: 0px; }

/*** anchor tags ***/
a:link, a:visited, a:hover, a:active { color: #CE534D; text-decoration: none; }

a:hover { text-decoration: underline; }

/*** main content list ***/
.main-list-item { font-weight: bold; font-size: 1.2em; margin: 0.8em 0; }

/* override the left margin added by font awesome for the main content list,
since it must be aligned with the content */
.fa-ul.main-list { margin-left: 0; }

/* list icons */
.main-list-item-icon { width: 36px; color: #46433A; }

/*** logo ***/
.logo-container { text-align: center; }

.logo { width: 96px; height: 96px; display: inline-block; background-size: cover; border-radius: 50%; -moz-border-radius: 50%; border: 2px solid #F1EED9; box-shadow: 0 0 0 3px #46433A; }

/*** author ***/
.author-container h1 { font-size: 2.8em; margin-top: 0; margin-bottom: 0; text-align: center; }

/*** tagline ***/
.tagline-container p { font-size: 1.3em; text-align: center; margin-bottom: 2em; }

/******/
hr { border: 0; height: 1px; background-image: -webkit-linear-gradient(left, transparent, #46433A, transparent); background-image: -moz-linear-gradient(left, transparent, #46433A, transparent); background-image: -ms-linear-gradient(left, transparent, #46433A, transparent); background-image: -o-linear-gradient(left, transparent, #46433A, transparent); }

/*** footer ***/
footer { position: fixed; bottom: 0; right: 0; height: 20px; }

.poweredby { font-family: "Arial Narrow", Arial; font-size: 0.6em; line-height: 0.6em; padding: 0 5px; }

/*** media queries ***/
/* X-Small devices (phones, 480px and up) */
@media (min-width: 480px) { /* wrapper stays 480px wide past 480px wide and is kept centered */
  .wrapper { width: 480px; margin: 10% auto 0 auto; } }
/* All other devices (768px and up) */
@media (min-width: 768px) { /* past 768px the layout is changed and the wrapper has a fixed width of 680px to accomodate both the header column and the content column */
  .wrapper { width: 680px; }
  /* the header column stays left and has a dynamic width with all contents aligned right */
  header { float: left; width: 46%; text-align: right; }
  .author-container h1, .logo-container, .tagline-container p { text-align: right; }
  main { width: 46%; margin-left: 54%; padding: 0; } }

  </style>


  <meta name="author" content="Zack Scholl">
  <meta name="description" content=""/>

  </head>
  <body>
  

  <div class="wrapper">
    <header>
        
        <div class="logo-container">
          <a class="logo" href="" style="background-image: url('/static/img/compass.png')"></a>
        </div>
        

        
        <div class="author-container"><h1>Just Read.</h1></div>
        

        
        <div class="tagline-container"><p>Copy and paste URL to get just the text content of that site.
	
		<br>
		<iframe src="https://ghbtns.com/github-btn.html?user=schollz&repo=justread&type=star&count=true&size=large" frameborder="0" scrolling="0" width="160px" height="30px"></iframe>
	</div>
        
    </header>
    <main>
      
      <div class="content">
               <form action='/' method='POST'>
                <input type='text' class="form-control input-lg" name='group' id='group' placeholder='Type a URL' autofocus></input>
               </form>
               <p class="lead">...or use /?url=X in the browser.</p>

      </div>
      
    </main>
</div>


</body>
</html>
`

	if r.Method == "GET" {
		defer timeTrack(time.Now(), "GET REQUEST")
		url := r.URL.Query()["url"]
		if len(url) > 0 {
			fmt.Fprintln(w, parseURL(url[0]))
		} else {
			fmt.Fprintln(w, indexPage)
		}
	} else {
		defer timeTrack(time.Now(), "POST REQUEST")
		r.ParseForm()
		url := r.Form["group"][0]
		fmt.Fprintln(w, parseURL(url))
	}
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
