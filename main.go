package main

import (
	"bufio"
	"crypto/tls"
	"fmt"
	"net/url"
	"os"
	"sync"
	"time"

	flags "github.com/jessevdk/go-flags"
	fasthttp "github.com/valyala/fasthttp"
)

// CLI options
var opts struct {
	Threads   int    `short:"t" long:"threads" default:"20" description:"Number of concurrent threads"`
	InputFile string `short:"i" long:"input" description:"Input file containing line seperated e-mail addresses, otherwise defaults to STDIN"`
	Domain    string `short:"d" long:"domain" default:"outlook.office365.com" description:"Autodiscover domain to use"`
	UserAgent string `short:"u" long:"user-agent" default:"Microsoft Office/16.0 (Windows NT 10.0; Microsoft Outlook 16.0.12026; Pro)" description:"User specified User agent to overwrite default"`
	Verbose   bool   `short:"v" long:"verbose" description:"Turns on verbose logging"`
	Insecure  bool   `long:"insecure" description:"Switches all HTTPS calls to HTTP"`
}

func main() {
	_, err := flags.ParseArgs(&opts, os.Args)
	if err != nil {
		panic(err)
	}

	//set amount of Threads
	numWorkers := opts.Threads
	work := make(chan string)
	go func() {
		// Default to reading standard in, but change if we specify input file
		scanner := bufio.NewScanner(nil)
		if opts.InputFile != "" {
			file, err := os.Open(opts.InputFile)
			if err != nil {
				fmt.Fprintln(os.Stderr, "ERR: could not load input file please try again", err)
				os.Exit(1)
			}
			scanner = bufio.NewScanner(file)
			defer file.Close()

		} else {
			scanner = bufio.NewScanner(os.Stdin)
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
	// create one webclient
	client := &fasthttp.Client{
		MaxConnsPerHost:               1024,
		DisableHeaderNamesNormalizing: true,
		MaxConnWaitTimeout:            20 * time.Second,
		ReadTimeout:                   3 * time.Second,
		NoDefaultUserAgentHeader:      true,
		TLSConfig:                     &tls.Config{InsecureSkipVerify: true},
	}

	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go doWork(work, wg, client) //Schedule the work
	}
	wg.Wait() //Wait for it all to complete
}

func doWork(work chan string, wg *sync.WaitGroup, wc *fasthttp.Client) {
	defer wg.Done()
	//It is unsafe using Request object from concurrently running goroutines, even for marshaling the request.
	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	req.Header.SetUserAgent(opts.UserAgent)
	req.Header.Set(fasthttp.HeaderAccept, "application/json")
	req.Header.SetMethod(fasthttp.MethodGet)
	resp.SkipBody = true
	prefix := "https://"
	if opts.Insecure {
		prefix = "http://"
	}
	for inputEmail := range work {
		req.SetRequestURI(fmt.Sprintf("%s%s/autodiscover/autodiscover.json/v1.0/%s%s", prefix, opts.Domain, inputEmail, "?Protocol=Autodiscoverv1"))
		err := wc.Do(req, resp) // do not follow redirect, just read
		if err != nil {
			fmt.Fprintln(os.Stderr, "ERR: performing request", err)
			fasthttp.ReleaseResponse(resp)
			continue
		}
		if loc := resp.Header.Peek("Location"); loc != nil {
			url, _ := url.Parse(string(loc))
			if url.RawQuery != "" {
				if email := url.Query()["Email"]; email == nil {
					if opts.Verbose { // Only print to stderr so we don't redirect it to a file
						fmt.Fprintln(os.Stderr, "ERR: Got redirect for", inputEmail, "but couldn't find Email parameter")
					}
				} else {
					fmt.Printf("%s, %s \n", inputEmail, email[0])
				}
			}
		}
		fasthttp.ReleaseResponse(resp)
	}
	fasthttp.ReleaseRequest(req)
}
