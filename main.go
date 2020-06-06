package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"sync"

	flags "github.com/jessevdk/go-flags"
)

// CLI options
var opts struct {
	Threads   int    `short:"t" long:"threads" default:"20" description:"Number of concurrent threads"`
	InputFile string `short:"i" long:"input" description:"Input file containing line seperated e-mail addresses, otherwise defaults to STDIN"`
	//Output	   	string `short:"o" long:"output" choice:"tcp" choice:"udp" default:"udp" description:"Protocol to use for lookups"`
	Domain    string `short:"d" long:"domain" default:"outlook.office365.com" description:"Autodiscover domain to use"`
	UserAgent string `short:"u" long:"user-agent" default:"Microsoft Office/16.0 (Windows NT 10.0; Microsoft Outlook 16.0.12026; Pro)" description:"User specified User agent to overwrite default"`
	Verbose   bool   `short:"v" long:"verbose" description:"Turns on verbose logging"`
	Insecure  bool   `long:"insecure" description:"Switches all HTTPS calls to HTTP"`
}

func main() {
	//Just a fancy ass banner
	asciiArt :=
		`
	██████╗  ██████╗ ███╗   ███╗██╗   ██╗
	██╔════╝ ██╔═══██╗████╗ ████║██║   ██║
	██║  ███╗██║   ██║██╔████╔██║██║   ██║
	██║   ██║██║   ██║██║╚██╔╝██║██║   ██║
	╚██████╔╝╚██████╔╝██║ ╚═╝ ██║╚██████╔╝
	╚═════╝  ╚═════╝ ╚═╝     ╚═╝ ╚═════╝ 
	`
	fmt.Println(asciiArt)
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "%s", err)
		os.Exit(1)
	}

	//set amount of Threads
	numWorkers := opts.Threads

	work := make(chan string)
	go func() {
		// Default to reading standard in, but change if we specify input file
		scanner := bufio.NewScanner(nil)
		if opts.InputFile != "" {
			file, _ := os.Open(opts.InputFile)
			defer file.Close()
			scanner = bufio.NewScanner(bufio.NewReader(file))
		} else {
			scanner = bufio.NewScanner(bufio.NewReader(os.Stdin))
		}
		// for each line on input
		for scanner.Scan() {
			work <- scanner.Text()
		}
		if err := scanner.Err(); err != nil {
			fmt.Fprintln(os.Stderr, "ERR: reading input failed:", err)
		}
		close(work)
	}()
	// Create a waiting group
	wg := &sync.WaitGroup{}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go doWork(work, wg) //Schedule the work
	}
	wg.Wait() //Wait for it all to complete
}

func doWork(work chan string, wg *sync.WaitGroup) {
	defer wg.Done()
	// Disable verifying of TLS certificates, unneccesary in tool's usecase
	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	//For control over HTTP client headers, redirect policy, and other settings, create a Client:
	client := &http.Client{
		Transport: customTransport,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse // do not follow redirects but also do not error
		},
	}
	prefix := "https://"
	if opts.Insecure {
		prefix = "http://"
	}
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s%s", prefix, opts.Domain), nil)
	req.Header.Add("User-Agent", opts.UserAgent)
	req.Header.Add("Accept", "application/json")
	req.URL.RawQuery = "Protocol=Autodiscoverv1"
	req.URL.ForceQuery = true
	req.Close = true
	for input_email := range work {
		req.URL.Path = url.PathEscape(fmt.Sprintf("/autodiscover/autodiscover.json/v1.0/%s", input_email))
		resp, err := client.Do(req)
		if err != nil {
			if opts.Verbose {
				fmt.Fprintln(os.Stderr, "ERR: performing request", err)
			}
			continue //TODO: Add Error-handling
		}
		loc, _ := resp.Location()
		if loc != nil {
			q := loc.Query()
			email := q["Email"]
			if email == nil {
				if opts.Verbose { // Only print to stderr so we don't redirect it to a file
					fmt.Fprintln(os.Stderr, "ERR: Got redirect for %s but couldn't find Email parameter", input_email)
				}
				continue
			}
			fmt.Printf("%s, %s \n", input_email, email[0])
		}
		continue
	}
}
