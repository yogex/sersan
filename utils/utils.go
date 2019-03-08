package utils

import (
	"encoding/json"
  "log"
	"net"
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/salestock/sersan/config"
)

var (
	num     uint64
	numLock sync.RWMutex
)

// Serial Serial
func Serial() uint64 {
	numLock.Lock()
	defer numLock.Unlock()
	id := num
	num++
	return id
}

// JsonError JSON error
func JsonError(w http.ResponseWriter, msg string, code int) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	json.NewEncoder(w).Encode(
		map[string]interface{}{
			"value": map[string]string{
				"message": msg,
			},
			"status": 13,
		})
}

// SecondsSince Calculate seconds since specified time until now
func SecondsSince(start time.Time) float64 {
	return float64(time.Now().Sub(start).Seconds())
}

// RequestInfo Request info
func RequestInfo(r *http.Request) (string, string) {
	user := ""
	if u, _, ok := r.BasicAuth(); ok {
		user = u
	} else {
		user = "unknown"
	}
	remote := r.Header.Get("X-Forwarded-For")
	if remote != "" {
		return user, remote
	}
	remote, _, _ = net.SplitHostPort(r.RemoteAddr)
	return user, remote
}

// GenerateSessionId Generate Session ID in JWT token format
func GenerateSessionID(sessionId, podName string, url *url.URL, baseURL string, vncPort int32) (formattedSessionId string, err error) {
	conf := config.Get()
	data := jwt.MapClaims{
		"sessionId": sessionId,
		"podName":   podName,
		"host":      url.Hostname(),
		"port":      url.Port(),
		"baseURL":   baseURL,
		"vncPort":   vncPort,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, data)
	formattedSessionId, err = token.SignedString([]byte(conf.SigningKey))
	if err != nil {
		log.Printf("Failed to create formatted session id %v", err)
		return
	}
	log.Printf("Generated Session ID %s", formattedSessionId)
	return
}

// ParseSessionID Extract session id information
func ParseSessionID(formattedSessionId string) (sessionId, podName, host, port, baseURL string, err error) {
	conf := config.Get()
	token, err := jwt.Parse(formattedSessionId, func(token *jwt.Token) (interface{}, error) {
		return []byte(conf.SigningKey), nil
	})
	if err != nil {
		log.Printf("Failed to parse session id %v", err)
		return
	}
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		sessionId = claims["sessionId"].(string)
		podName = claims["podName"].(string)
		host = claims["host"].(string)
		port = claims["port"].(string)
		baseURL = claims["baseURL"].(string)
	}
	return
}
