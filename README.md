Pepper
======

_Personal web-proxy written in Go._

Pepper was developed to fix two problems:

1. Improve privacy while surfing the web
2. Unify keyword search among all browsers

The first problem is tackled by black- and whitelising domains known to track users and blocking web-bugs like single pixel GIFs. The second problem is more of a convenience and allows you to setup keyword search once and use them on all browsers.

_This project is a work-in-process._

Features:

- Block web-bugs
- Blacklist domains known to track users and server ads
- Unify Keyword search engine among all browsers


Build
-----

To build Pepper you must have [Go][] installed.

[Go]: http://golang.org

```go
export GOPATH=$(pwd)/go
mkdir -p $GOPATH
go get github.com/namsral/pepper
go build pepper.go sem.go
```

Install
-------

### OSX

On OSX you can have pepper run as a service by creating a launchagent.

The following setup assumes you installed the pepper binary at `/usr/local/bin/pepper` and the data file at `/Users/namsral/pepper.json`.

1. Create a new agent:

```sh
cat > ~/Library/LaunchAgents/com.namsral.pepper.plist
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
   <key>Label</key>
   <string>com.namsral.pepper</string>
   <key>ProgramArguments</key>
   <array>
      <string>-data=/usr/local/bin/pepper</string>
      <string>-data=/Users/namsral/pepper.json</string>
      <string>-http=127.0.0.1:8081</string>
   </array>
   <key>Sockets</key>
   <dict>
      <key>Listeners</key>
      <dict>
         <key>SockServiceName</key>
         <string>8081</string>
      </dict>
   </dict>
   <key>KeepAlive</key>
   <true/>
   </dict>
</plist>
```

2. Enable the new agent and have Pepper started automatically:

```
sudo launchctl load -F ~/Library/LaunchAgents/com.namsral.pepper.plist
```
 
### *Nix

On other *nixes run pepper in the background:

The following setup assumes you installed the pepper binary at `/usr/local/bin/pepper` and the data file at `/home/namsral/pepper.json`.

```
./pepper -http=127.0.0.1:8081 -data=~/pepper.json
```

_Or you could write a startup file for your favourite init system._


### Browser setup

1. Launch your browser and configure the HTTP proxy to point to host:127.0.0.1 and port:8081
2. Surf to `http://pepper` to check the new proxy is working
3. If your browser support it add a custom search engine with url: `http://pepper/search?q=%s`, or set your search engine to Bing as Pepper wil Hijack any search request made to `http://www.bing.com/search`.
4. One last thing to do is edit pepper.json and add your own domain blacklist and search engine keywords.


Todo
----

As mentioned above, this is a work-in-process project so there are things todo:

- Persist blacklisted web-bug URLs
- Move to a datastore
- More efficient response handling, preferebly within the proxyHandler
- Expand webBugHandler to block other tracking methods beside single pixel GIFs
- Modify resources from the browser; search engines, blacklists, etc 
- Multiuser support; home or small office use

License
-------

Pepper is licensed under the terms of the MIT license, see attached LICENSE file for more details.
