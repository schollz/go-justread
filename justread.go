package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"time"
	"regexp"
)

// https://play.golang.org/p/N8A5cXa3RM
func wordCountWithoutLinks(data string) (int) {
// from https://regex101.com/r/tZ6yK9/3
	re := regexp.MustCompile("\\[([^\\]]+)\\]\\(([^\\)]+)\\)")
	data := "#### [POLITICO Florida Playbook: The Spies Effect – Rubio’s Obamacare ‘sabotage’ – Meyer Lansky’s fam wants Cuba compensation – Charter school lobby’s hardball – Help for the Florida Center for Investigative Reporting – The mockingbird’s NRA firepower](http://www.politico.com/tipsheets/florida-playbook/2015/12/politico-florida-playbook-the-spies-effect-rubios-obamacare-sabotage-meyer-lanskys-fam-wants-cuba-compensation-charter-school-lobbys-hardball-help-for-the-florida-center-for-investigative-reporting-the-mockingbirds-nra-firepower-211676)"
	matches := re.FindAllString(data, -1)
	fmt.Println(matches)
	newdata := data
	for _, match := range matches {
		if match[0:2] == "[!" {
			match = match[2:]
		}
		newdata = strings.Replace(newdata, match, "", -1)
	}
	fmt.Println(newdata)
	data = newdata
	matches = re.FindAllString(data, -1)
	fmt.Println(matches)
	newdata = data
	for _, match := range matches {
		if match[0:3] == "[![" {
			match = match[2:]
		}
		newdata = strings.Replace(newdata, match, "", -1)
	}
	fmt.Println(newdata)
	return len(strings.Split(newdata, " "))
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


func printHTML() {
	var (
		cmdOut []byte
		err    error
	)
	cmdName := "pandoc"
	cmdArgs := []string{"test.md", "-t", "html"}
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running git rev-parse command: ", err)
		os.Exit(1)
	}
	sha := string(cmdOut)
	fmt.Printf(sha)
}

func main() {
	defer timeTrack(time.Now(), "wget")
	var (
		cmdOut []byte
		err    error
	)
	cmdName := "pandoc"
	cmdArgs := []string{"--columns=700", "-t", "markdown", "-s", "-r", "html", "http://www.newyorker.com/culture/culture-desk/bill-murrays-little-christmas-miracle?mbid=rss"}
	if cmdOut, err = exec.Command(cmdName, cmdArgs...).Output(); err != nil {
		fmt.Fprintln(os.Stderr, "There was an error running git rev-parse command: ", err)
		os.Exit(1)
	}
	sha := string(cmdOut)
	defer timeTrack(time.Now(), "parsing")
	result := strings.Split(sha, "\n")

	// Display all elements.
	start := 0
	end := 0
	longest := 0
	bestStart := 0
	bestEnd := 0
	strikes := 0
	totalWords := 0
	for i := range result {
		strLen := len(strings.Split(result[i], " "))
		totalWords += strLen
		if (strLen) > 1 {
			if (strLen) > 10 {
				if start > 0 {
					end = i + 1
				} else {
					start = i - 1
				}
			} else if strikes > 0 {
				if totalWords > longest {
					longest = totalWords
					bestStart = start
					bestEnd = end
				}
				fmt.Printf("\n\nTotal words: %d\n\n", totalWords)
				start = 0
				end = 0
				strikes = 0
				totalWords = 0
			} else {
				strikes++
			}
		}
		fmt.Printf("%d) %d", i, strLen)
		fmt.Println(result[i])

	}

	fmt.Printf("%d) %d", bestStart, bestEnd)
	fmt.Println("BEST RESULT:\n\n")
	f, err := os.Create("test.md")
	check(err)
	defer f.Close()
	w := bufio.NewWriter(f)
	lastSentenceGood := false
	for i := range result {
		strLen := len(strings.Split(result[i], " "))
		strLen2 := len(strings.TrimSpace(result[i]))

		if strLen > 20 || strLen2 == 0 {
			fmt.Printf("%s\n", result[i])
			w.WriteString(result[i])
			w.WriteString("\n")
			if strLen > 20 {
				lastSentenceGood = true
			}
		} else if lastSentenceGood == true && strLen > 7 {
			fmt.Printf("%s\n", result[i])
			w.WriteString(result[i])
			w.WriteString("\n")
			lastSentenceGood = false
		} else if strLen > 1 {
			lastSentenceGood = false
		}
	}
	w.Flush()
	fmt.Println("WEB RESULT:\n\n")
	printHTML()
}
