package httpserver

import (
	"html/template"
	"net/http"
	"path"

	"github.com/aWZHY0yQH81uOYvH/goshs/logger"
	"github.com/aWZHY0yQH81uOYvH/goshs/utils"
)

func (fs *FileServer) handleError(w http.ResponseWriter, req *http.Request, err error, status int) {
	// Set header to status
	w.WriteHeader(status)

	// Define empty error
	var e httperror

	// Log to console
	logger.LogRequest(req, status, fs.Verbose)

	// Construct error for template filling
	e.ErrorCode = status
	e.ErrorMessage = err.Error()
	e.AbsPath = path.Join(fs.Webroot, req.URL.Path)
	e.GoshsVersion = fs.Version

	// Template handling
	file, err := static.ReadFile("static/templates/error.html")
	if err != nil {
		logger.Errorf("opening embedded file: %+v", err)
	}
	t := template.New("error")
	if _, err := t.Parse(string(file)); err != nil {
		logger.Errorf("parsing the template: %+v", err)
	}
	if err := t.Execute(w, e); err != nil {
		logger.Errorf("executing the template: %+v", err)
	}
}

func (fs *FileServer) logStart(what string) {
	var interfaceAdresses map[string]string
	var err error
	if what == modeWeb {
		if fs.IP == "0.0.0.0" {
			interfaceAdresses, err = utils.GetAllIPAdresses()
			if err != nil {
				logger.Errorf("There has been an error fetching the interface addresses: %+v\n", err)
			}
			for k, v := range interfaceAdresses {
				logger.Infof("Serving on interface %s bound to %s:%+v\n", k, v, fs.Port)
			}
		} else {
			logger.Infof("Serving on %s:%+v\n", fs.IP, fs.Port)
		}
	}

	protocol := "HTTP"
	if fs.SSL {
		protocol = "HTTPS"
	}

	switch what {
	case modeWeb:
		if fs.SSL {
			// Check if selfsigned
			if fs.SelfSigned {
				logger.Infof("Serving %s from %+v with ssl enabled and self-signed certificate\n", protocol, fs.Webroot)
				logger.Warn("Be sure to check the fingerprint of certificate")
				logger.Infof("SHA-256 Fingerprint: %+v\n", fs.Fingerprint256)
				logger.Infof("SHA-1   Fingerprint: %+v\n", fs.Fingerprint1)
			} else {
				logger.Infof("Serving %s from %+v with ssl enabled server key: %+v, server cert: %+v\n", protocol, fs.Webroot, fs.MyKey, fs.MyCert)
				logger.Info("You provided a certificate and might want to check the fingerprint nonetheless")
				logger.Infof("SHA-256 Fingerprint: %+v\n", fs.Fingerprint256)
				logger.Infof("SHA-1   Fingerprint: %+v\n", fs.Fingerprint1)
			}
		} else {
			logger.Infof("Serving %s from %+v\n", protocol, fs.Webroot)
		}
	case "webdav":
		if fs.SSL {
			// Check if selfsigned
			if fs.SelfSigned {
				logger.Infof("Serving WEBDAV on %+v:%+v from %+v with ssl enabled and self-signed certificate\n", fs.IP, fs.WebdavPort, fs.Webroot)
				logger.Warn("WARNING! Be sure to check the fingerprint of certificate")
				logger.Infof("SHA-256 Fingerprint: %+v\n", fs.Fingerprint256)
				logger.Infof("SHA-1   Fingerprint: %+v\n", fs.Fingerprint1)
			} else {
				logger.Infof("Serving WEBDAV on %+v:%+v from %+v with ssl enabled server key: %+v, server cert: %+v\n", fs.IP, fs.WebdavPort, fs.Webroot, fs.MyKey, fs.MyCert)
				logger.Info("INFO! You provided a certificate and might want to check the fingerprint nonetheless")
				logger.Infof("SHA-256 Fingerprint: %+v\n", fs.Fingerprint256)
				logger.Infof("SHA-1   Fingerprint: %+v\n", fs.Fingerprint1)
			}
		} else {
			logger.Infof("Serving WEBDAV on %+v:%+v from %+v\n", fs.IP, fs.WebdavPort, fs.Webroot)
		}
	default:
	}
}
