package session

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
	"time"

	cache "github.com/patrickmn/go-cache"
	"github.com/salestock/sersan/config"
	"github.com/salestock/sersan/utils"
)

const slash = "/"

var (
	httpClient = &http.Client{
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	hostname string
)

// SessionHandler Session handler
type SessionHandler struct {
	SessionService *SessionService `inject:""`
	TunedTransport *http.Transport `inject:""`
	Cache          *cache.Cache    `inject:""`
}

// Create Handler for new session request
func (h SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
	sessionStartTime := time.Now()
	hostname, err := os.Hostname()
	if err != nil {
		log.Printf("Init %s: %v", os.Args[0], err)
	}
	conf := config.Get()
	user, remote := utils.RequestInfo(r)
	body, err := ioutil.ReadAll(r.Body)
	r.Body.Close()
	if err != nil {
		log.Printf("Error Reading Request %v", err)
		return
	}
	var browser *Browser
	err = json.Unmarshal(body, &browser)
	if err != nil {
		log.Printf("Error Reading Request %v", err)
		return
	}
	if browser.W3CCaps.Caps.Name != "" && browser.Caps.Name == "" {
		browser.Caps = browser.W3CCaps.Caps
	}

	gridStarter, _ := h.SessionService.Create(browser)
	startedGrid, err := gridStarter.StartWithCancel()
	if err != nil {
		log.Printf("Failed to create pod: %v", err)
		return
	}

	var resp *http.Response
	i := 1
	for ; ; i++ {
		r.URL.Host, r.URL.Path = startedGrid.URL.Hostname()+":"+startedGrid.URL.Port(), startedGrid.Grid.Grid.BaseURL+"/session"
		log.Printf("Request URL: %s", r.URL.String())
		req, _ := http.NewRequest(http.MethodPost, r.URL.String(), bytes.NewReader(body))
		ctx, done := context.WithTimeout(r.Context(), 60*time.Second)
		defer done()
		log.Printf("Session attempted to %s for %d time{s)", startedGrid.URL.Hostname(), i)
		rsp, err := httpClient.Do(req.WithContext(ctx))
		select {
		case <-ctx.Done():
			if rsp != nil {
				rsp.Body.Close()
			}
			switch ctx.Err() {
			case context.DeadlineExceeded:
				log.Printf("Session attempted timeout after %d times", conf.NewSessionAttemptTimeout)
				if int32(i) < conf.RetryCount {
					log.Printf("Retry count %d", conf.RetryCount)
					continue
				}
				err := fmt.Errorf("New session attempts retry count exceeded")
				log.Printf("Session for %s failed: %s", startedGrid.URL.Hostname(), err)
				utils.JsonError(w, err.Error(), http.StatusInternalServerError)
			case context.Canceled:
				log.Printf("Client disconnected %s - %s - %.2fs", user, remote, utils.SecondsSince(sessionStartTime))
			}
			startedGrid.Cancel()
			return
		default:
		}
		if err != nil {
			if rsp != nil {
				rsp.Body.Close()
			}
			log.Printf("Session failed %s", err)
			startedGrid.Cancel()
			utils.JsonError(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if rsp.StatusCode == http.StatusNotFound {
			continue
		}
		resp = rsp
		break
	}
	defer resp.Body.Close()
	var reply map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&reply)
	var sessionID string
	// Webdriver response
	if reply["sessionId"] != nil {
		sessionID = reply["sessionId"].(string)
	}

	// Appium response
	if reply["value"] != nil {
		if reply["value"].(map[string]interface{})["sessionId"] != nil {
			sessionID = reply["value"].(map[string]interface{})["sessionId"].(string)
		}
	}

	log.Printf("Session ID: %s", sessionID)
	sessionInfo := &utils.SessionInfo{
		SessionID:   sessionID,
		ServiceName: startedGrid.Name,
		Host:        startedGrid.URL.Hostname(),
		Port:        startedGrid.URL.Port(),
		BaseURL:     startedGrid.Grid.Grid.BaseURL,
		VNCPort:     fmt.Sprintf("%d", startedGrid.Grid.Grid.VNCPort),
		Engine:      startedGrid.Grid.Grid.Engine,
	}
	proxy := &httputil.ReverseProxy{
		Transport: h.TunedTransport,
	}
	cacheInfo := &utils.CachedInfo{
		Session: sessionInfo,
		Proxy:   proxy,
	}

	formattedSessionID, err := utils.GenerateSessionID(sessionInfo)
	h.Cache.Set(formattedSessionID, cacheInfo, cache.DefaultExpiration)
	if err != nil {
		log.Printf("Failed to get formatted session id: %v", err)
		return
	}
	reply["sessionId"] = formattedSessionID
	json.NewEncoder(w).Encode(reply)
	var s struct {
		Value struct {
			ID string `json:"sessionId"`
		}
		ID string `json:"sessionId"`
	}
	location := resp.Header.Get("Location")
	if location != "" {
		l, err := url.Parse(location)
		log.Printf("Location: %v", l)
		if err == nil {
			fragments := strings.Split(l.Path, slash)
			s.ID = fragments[len(fragments)-1]
			u := &url.URL{
				Scheme: "http",
				Host:   hostname,
				Path:   path.Join(startedGrid.Grid.Grid.BaseURL+"/session", s.ID),
			}
			w.Header().Add("Location", u.String())
			w.WriteHeader(resp.StatusCode)
		}
	} else {
		tee := io.TeeReader(resp.Body, w)
		w.WriteHeader(resp.StatusCode)
		json.NewDecoder(tee).Decode(&s)
		if s.ID == "" {
			s.ID = sessionID
		}
		log.Printf("Location empty %v", s.Value.ID)
	}
	if s.ID == "" {
		log.Printf("Session failed %s", resp.Status)
		startedGrid.Cancel()
		return
	}

	log.Printf("Session created with id %s %d in %.2fs", s.ID, i, utils.SecondsSince(sessionStartTime))
}

// Proxy Handler for all incoming request other than new session
func (h SessionHandler) Proxy(w http.ResponseWriter, r *http.Request) {
	done := make(chan func())
	fragments := strings.Split(r.URL.Path, slash)
	sessionID := fragments[2]
	var sessionInfo *utils.SessionInfo
	var proxy *httputil.ReverseProxy
	cachedInfo, found := h.Cache.Get(sessionID)
	if found {
		log.Printf("Found cached session ID %s", sessionID)
		sessionInfo = cachedInfo.(*utils.CachedInfo).Session
		proxy = cachedInfo.(*utils.CachedInfo).Proxy
	} else {
		log.Printf("Parse session ID %s", sessionID)
		s, err := utils.ParseSessionID(sessionID)
		if err != nil {
			log.Printf("Invalid session ID %s", sessionID)
		}
		sessionInfo = s
		proxy = &httputil.ReverseProxy{
			Transport: h.TunedTransport,
		}
	}
	go func(w http.ResponseWriter, r *http.Request) {
		cancel := func() {}
		defer func() {
			done <- cancel
		}()
		proxy.Director = func(r *http.Request) {
			fragments[2] = sessionInfo.SessionID
			r.URL.Path = path.Join(sessionInfo.BaseURL+slash, strings.Join(fragments, slash))
			r.URL.Host = sessionInfo.Host + ":" + sessionInfo.Port
		}
		proxy.ServeHTTP(w, r)
		if r.Method == http.MethodDelete && len(fragments) == 3 {
			defer func() {
				err := h.SessionService.Delete(sessionInfo.ServiceName, sessionInfo.Engine)
				if err != nil {
					log.Printf("Unable to delete pod %s", sessionInfo.ServiceName)
				}
			}()
		}
	}(w, r)
	go (<-done)()
}
