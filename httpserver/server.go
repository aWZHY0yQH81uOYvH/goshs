package httpserver

import (
	"fmt"
	"net"
	"net/http"
	"time"
	"crypto/tls"

	"github.com/gorilla/mux"
	"github.com/aWZHY0yQH81uOYvH/goshs/ca"
	"github.com/aWZHY0yQH81uOYvH/goshs/clipboard"
	"github.com/aWZHY0yQH81uOYvH/goshs/logger"
	"github.com/aWZHY0yQH81uOYvH/goshs/ws"
	"golang.org/x/net/webdav"
)

// Start will start the file server
func (fs *FileServer) Start(what string) {
	var addr string
	// Setup routing with gorilla/mux
	mux := mux.NewRouter()

	switch what {
	case modeWeb:
		mux.Methods(http.MethodPost).HandlerFunc(fs.upload)
		mux.PathPrefix("/").HandlerFunc(fs.handler)

		addr = fmt.Sprintf("%+v:%+v", fs.IP, fs.Port)
	case "webdav":
		wdHandler := &webdav.Handler{
			FileSystem: webdav.Dir(fs.Webroot),
			LockSystem: webdav.NewMemLS(),
			Logger: func(r *http.Request, e error) {
				if e != nil && r.Method != "PROPFIND" {
					logger.Errorf("WEBDAV: %s - - \"%s %s %s\"", r.RemoteAddr, r.Method, r.URL.Path, r.Proto)
					return
				} else if r.Method != "PROPFIND" {
					logger.Infof("WEBDAV:  %s - - \"%s %s %s\"", r.RemoteAddr, r.Method, r.URL.Path, r.Proto)
				}
			},
		}

		mux.PathPrefix("/").Handler(wdHandler)
		addr = fmt.Sprintf("%+v:%+v", fs.IP, fs.WebdavPort)
	default:
	}

	// construct and bind listener
	listener, err := net.Listen("tcp", addr)
	if err != nil {
		logger.Fatalf("Error binding to listener '%s': %+v", addr, err)
	}
	defer listener.Close()

	// construct server
	server := http.Server{
		// Addr:              addr,
		Handler:           http.AllowQuerySemicolons(mux),
		ReadHeaderTimeout: 10 * time.Second, // Mitigate Slow Loris Attack
		// Against good practice no timeouts here, otherwise big files would be terminated when downloaded
	}

	// init clipboard
	fs.Clipboard = clipboard.New()

	// init websocket hub
	fs.Hub = ws.NewHub(fs.Clipboard)
	go fs.Hub.Run()

	// Check BasicAuth and use middleware
	if (fs.User != "" || fs.Pass != "") && what == modeWeb {
		if !fs.SSL {
			logger.Warnf("You are using basic auth without SSL. Your credentials will be transferred in cleartext. Consider using -s, too.")
		}
		logger.Infof("Using basic auth with user '%s' and password '%s'", fs.User, fs.Pass)
		// Use middleware
		mux.Use(fs.BasicAuthMiddleware)
	}

	if fs.Silent {
		logger.Info("Serving in silent mode - no dir listing available at HTTP Listener")
	}

	// Check if ssl
	if fs.SSL {
		// Check if selfsigned
		if fs.SelfSigned {
			serverTLSConf, fingerprint256, fingerprint1, err := ca.Setup()
			if err != nil {
				logger.Fatalf("Unable to start SSL enabled server: %+v\n", err)
			}
			server.TLSConfig = serverTLSConf
			fs.Fingerprint256 = fingerprint256
			fs.Fingerprint1 = fingerprint1
			fs.logStart(what)

			// Drop privs if needed
			fs.dropPrivs()

			logger.Panic(server.ServeTLS(listener, "", ""))
			// logger.Panic(server.ListenAndServeTLS("", ""))
		} else {
			if fs.MyCert == "" || fs.MyKey == "" {
				logger.Fatal("You need to provide server.key and server.crt if -s and not -ss")
			}

			fingerprint256, fingerprint1, err := ca.ParseAndSum(fs.MyCert)
			if err != nil {
				logger.Fatalf("Unable to start SSL enabled server: %+v\n", err)
			}

			cert, err := tls.LoadX509KeyPair(fs.MyCert, fs.MyKey)
			if err != nil {
				logger.Fatalf("Failed to load provided key or certificate: %+v\n", err)
			}

			server.TLSConfig = &tls.Config{
				Certificates: []tls.Certificate{cert},
				MinVersion:   tls.VersionTLS12,
			}

			fs.Fingerprint256 = fingerprint256
			fs.Fingerprint1 = fingerprint1
			fs.logStart(what)

			// Drop privs if needed
			fs.dropPrivs()

			logger.Panic(server.ServeTLS(listener, "", ""))
			//logger.Panic(server.ListenAndServeTLS(fs.MyCert, fs.MyKey))
		}
	} else {
		fs.logStart(what)

		// Drop privs if needed
		fs.dropPrivs()

		logger.Panic(server.Serve(listener))
		//logger.Panic(server.ListenAndServe())
	}
}
