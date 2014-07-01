package main

import (
    "net/http"
    "net/http/httputil"
    "log"
    "github.com/namsral/flag"
    "strings"
    "fmt"
    "os"
    "encoding/json"
    "image/gif"
    "html/template"
    "sync/atomic"
)

type Data struct {
    Whitelist []string
    Blacklist []string
    Engines []SearchEngine
}

// webBugHandler blocks and blacklist web-bugs
// This handler will make the request and based on the
// response it will blacklist the URL if it finds a web-bug
// or hijack the request and serve a pixel of its own.
func webBugHandler(handler http.Handler) http.Handler {
    url_blacklist := []string{}
    fn := func(w http.ResponseWriter, r *http.Request) {

        urlStr := fmt.Sprintf("%s://%s%s", r.URL.Scheme, r.URL.Host, r.URL.Path)

        // Block web bug if it has been requested befored
        for _, u := range url_blacklist {
            if u == urlStr {
                log.Printf("INFO redirect-webbug %s", urlStr)
                atomic.AddUint64(&webBugCounter, 1)
                http.Redirect(w, r, "http://pepper/pixel.gif", 302)
                return
            }
        }

        // Get the response
        resp, err := http.DefaultClient.Get(r.RequestURI)
        if err != nil {
            log.Printf("ERROR PROXY %v", err.Error())
            handler.ServeHTTP(w, r)
            return
        }
        defer resp.Body.Close()

        // Add Web Bug URL to url url_blacklist
        if resp.ContentLength < 100 && resp.Header.Get("Content-Type") == "image/gif" {
            img, err:= gif.DecodeConfig(resp.Body)
            if err != nil {
                log.Printf("ERROR %v\n", err)
                handler.ServeHTTP(w, r)
                return
            }
            if img.Width * img.Height == 1 {
                log.Printf("INFO blacklisted-url %s", urlStr)
                url_blacklist = append(url_blacklist, urlStr)
            }
        }

        handler.ServeHTTP(w, r)
    }
    return http.HandlerFunc(fn)
}

// Hijack a call to a specific URI and redirect it to our own search engine
// Useful for the Safari browser as it doens't allow custom search engines
func hijackSearchEngine(handler http.Handler, uriprefix string) http.Handler {
    fn := func(w http.ResponseWriter, r *http.Request) {
        if strings.HasPrefix(r.RequestURI, uriprefix) {
            log.Printf("INFO hijacked-search-request %s", r.RequestURI)
            q := r.URL.Query().Get("q")
            urlStr := fmt.Sprintf("http://pepper/search?q=%s", q)
            http.Redirect(w, r, urlStr, 302)
            return
        }
        handler.ServeHTTP(w, r)
    }
    return http.HandlerFunc(fn)
}

func searchHandler() http.Handler {
    sem = NewSEManager()
    sem.Engines = data.Engines

    fn := func(w http.ResponseWriter, r *http.Request) {
        if r.Method != "GET" {
            http.StatusText(405)
            return
        }

        // Query
        query := r.URL.Query().Get("q")
        if len(query) > 0 {
            urlStr, err := sem.Query(query)
            if err != nil {
                http.Error(w, "Internal Server Error", 500)
                return
            }
            http.Redirect(w, r, urlStr, 302)
        }

        // Template
        webBugCounterState := atomic.LoadUint64(&webBugCounter)
        t, err := template.New("search").Parse(`<!DOCTYPE html>
<html>
<head>
    <meta charset="UTF-8" /> 
    <title>Pepper</title>
    <style type="text/css">
    body {color:#333;}
    input[type=text] {border:1px solid rgb(62,139,208);padding:8px 14px;line-height:15px;font-size:1.4rem;}
    input {margin:1em 0;}
    .searchbox {width:50%;margin:30% auto 70% auto;text-align:center;}
    </style>
</head>
<body onload="document.getElementById('q').focus();">
    <div class="content">
        <div><h2>Total web-bugs blocked: {{ .Counter }}<h2></div>
        <div class="searchbox">
            <form method="GET" action="/search">
                <input type="text" name="q" id="q"><br>
                <input type="submit" value="Search">
            </form>
        </div>
    </div>
</body>
</html>`)
        if err != nil {
            log.Fatal("Failed to create template", err)
        }
        d := map[string]string{"Counter": fmt.Sprintf("%d", webBugCounterState)}

        if err := t.ExecuteTemplate(w, "search", d); err != nil {
            log.Fatal("Failed to render template", err)
        }
    }
    return http.HandlerFunc(fn)
}


// Checks if the hostname contains the given domain
func dnsDomainIs(hostname, domain string) bool {
    // Match exact
    if domain == hostname {
        return true
    }
    
    // Match root domains: .example.com ~ example.com
    if len(domain) > 1 && strings.HasPrefix(domain, ".") && domain[1:] == hostname {
        return true
    }

    // Match subdomains; .example.com ~ ad.example.com
    if strings.HasPrefix(domain, ".") && strings.HasSuffix(hostname, domain) {
        return true
    }
    
    return false
}

func isHostAllowed(hostname string) bool {
    for _, s := range data.Whitelist {
        if dnsDomainIs(hostname, s) {
            log.Println("INFO whitelisted-domain", hostname)
            return true
        }
    }
    for _, s := range data.Blacklist {
        if dnsDomainIs(hostname, s) {
            log.Println("INFO blacklisted-domain", hostname)
            return false
        }
    }
    log.Println("INFO domain-allowed", hostname)
    return true
}

// domainFilterHandler will allow or block request to listed
// domain names.
func domainFilterHandler(handler http.Handler) http.Handler {
    fn := func(w http.ResponseWriter, r *http.Request) {
        // Accept whitelisted domains
        if isHostAllowed(r.Host) == true {
            handler.ServeHTTP(w, r)
            return
        }

        // Block blacklisted domains
        if isHostAllowed(r.Host) != true {
            w.WriteHeader(403)
            return
        }
        // Otherwise pass on to next handler
        handler.ServeHTTP(w, r)
    }
    return http.HandlerFunc(fn)
}

// logHandler is used for developing
func logHandler(handler http.Handler) http.Handler {
    fn := func(w http.ResponseWriter, r *http.Request) {
        log.Printf("INFO %s", r.RequestURI)
        handler.ServeHTTP(w, r)
    }
    return http.HandlerFunc(fn)
}

// proxyHandler makes the request to the destination and responds
// to the client.
func proxyHandler() http.Handler {
    fn := func(w http.ResponseWriter, r *http.Request) {
        proxy := &httputil.ReverseProxy{Director: func(req *http.Request) {
            req.URL = r.URL
        }}
        proxy.ServeHTTP(w, r)
    }
    return http.HandlerFunc(fn)
}

// htdocsHandler is used for developing
func htdocsHandler(handler http.Handler) http.Handler {
    fn := func(w http.ResponseWriter, r *http.Request) {
        if r.Host == *httpAddr {
            h := http.StripPrefix("/htdocs/", http.FileServer(http.Dir("./htdocs")))
            h.ServeHTTP(w, r)
            return
        }
        handler.ServeHTTP(w, r)
    }
    return http.HandlerFunc(fn)
}

// pixelHandler serves a single pixel GIF
func pixelHandler() http.Handler {
    fn := func(w http.ResponseWriter, r *http.Request) {
        gifBytes := []byte{
            0x47, 0x49, 0x46, 0x38, 0x39, 0x61, 0x01, 0x00, 0x01, 0x00, 0x80, 0xFF,
            0x00, 0xFF, 0xFF, 0xFF, 0x00, 0x00, 0x00, 0x2C, 0x00, 0x00, 0x00, 0x00,
            0x01, 0x00, 0x01, 0x00, 0x00, 0x02, 0x02, 0x44, 0x01, 0x00, 0x3B}
        w.Header().Set("Content-Type", "image/gif")
        w.Header().Set("Content-Length", "35")
        w.Header().Set("Cache-Control", "max-age=604800")
        w.Header().Set("X-Pixel", "Enjoy my pixel")
        w.Write(gifBytes)
    }
    return http.HandlerFunc(fn)
}

var (
    httpAddr = flag.String("http", "127.0.0.1:8080", "HTTP server listen address")
    dataPath = flag.String("data", "", "path to file containing data")
    sem *SEManager
    data Data
    webBugCounter uint64 = 0
)

func main() {
    flag.Parse()

    // Load data from JSON file
    data := Data{}
    fp, err := os.Open(*dataPath)
    if err != nil {
        fmt.Printf("ERROR read-file %v\n", err)
    }
    dec := json.NewDecoder(fp)
    if err := dec.Decode(&data); err != nil {
        fmt.Println("ERROR JSON %v\n", err)
    }

    mux := http.NewServeMux()
    
    // Handle proxy request
    mux.Handle("/", proxyHandler())

    // Serve single pixel GIFs
    mux.Handle("pepper/pixel.gif", pixelHandler())

    // Handle Search
    mux.Handle("pepper/search", searchHandler())

    // Redirect homepage for now
    mux.Handle("pepper/", http.RedirectHandler("http://pepper/search", 302))

    // Add middleware
    muxWithMiddlewares := webBugHandler(mux)
    muxWithMiddlewares = domainFilterHandler(muxWithMiddlewares)
    muxWithMiddlewares = hijackSearchEngine(muxWithMiddlewares, "http://www.bing.com/search")

    log.Printf("Running on http://%s", *httpAddr)

    if err := http.ListenAndServe(*httpAddr, muxWithMiddlewares); err != nil {
        log.Fatal(err)
    }
}