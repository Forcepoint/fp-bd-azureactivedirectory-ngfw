package elements

import (
	"github.com/sirupsen/logrus"
	"net/http"
)

func loggerWithField(r *http.Request) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"RequestMethod": r.Method, "RequestURL": r.RequestURI, "RemoteAddress": r.RemoteAddr,
	})
}

func LogFaTal(err error, endpoint string) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"errorMessage": err.Error(), "RequestURL": endpoint})
}

func LogInfo(endpoint string) *logrus.Entry {
	return logrus.WithFields(logrus.Fields{
		"RequestURL": endpoint})
}
