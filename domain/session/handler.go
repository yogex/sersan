package session

import (
  "bytes"
  "context"
  "encoding/json"
  "fmt"
  "log"
  "io"
  "io/ioutil"
  "net/http"
  "net/http/httputil"
  "net/url"
  "os"
  "path"
  "strings"
  "time"

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
}

// Create Handler for new session request
func (h SessionHandler) Create(w http.ResponseWriter, r *http.Request) {
  sessionStartTime := time.Now()
  hostname, err := os.Hostname()
  if err != nil {
    log.Printf("Init %s: %v", os.Args[0], err)
  }
  conf := config.Get()
  requestId := utils.Serial()
  user, remote := utils.RequestInfo(r)
  body, err := ioutil.ReadAll(r.Body)
  r.Body.Close()
  if err != nil {
    log.Printf("[%d] Error Reading Request %v", requestId, err)
    return
  }
  var browser *Browser
  err = json.Unmarshal(body, &browser)
  if err != nil {
    log.Printf("[%d] Error Reading Request %v", requestId, err)
    return
  }
  if browser.W3CCaps.Caps.Name != "" && browser.Caps.Name == "" {
    browser.Caps = browser.W3CCaps.Caps
  }

  gridStarter, _ := h.SessionService.Create(browser, requestId)
  startedGrid, err := gridStarter.StartWithCancel()
  if err != nil {
    log.Printf("Failed to create pod: %v", err)
    return
  }

  var resp *http.Response
  i := 1
  for ; ; i++ {
    r.URL.Host, r.URL.Path = startedGrid.URL.Hostname()+":"+startedGrid.URL.Port(), startedGrid.Grid.Grid.BaseURL+"/session"
    log.Printf("%s", r.URL.String())
    req, _ := http.NewRequest(http.MethodPost, r.URL.String(), bytes.NewReader(body))
    ctx, done := context.WithTimeout(r.Context(), 60*time.Second)
    defer done()
    log.Printf("[%d] Session attempted - %s - %d", requestId, startedGrid.URL.Hostname(), i)
    rsp, err := httpClient.Do(req.WithContext(ctx))
    select {
    case <-ctx.Done():
      if rsp != nil {
        rsp.Body.Close()
      }
      switch ctx.Err() {
      case context.DeadlineExceeded:
        log.Printf("[%d] Session attempted timeout %d", requestId, conf.NewSessionAttemptTimeout)
        if int32(i) < conf.RetryCount {
          log.Printf("Retry count %d", conf.RetryCount)
          continue
        }
        err := fmt.Errorf("New session attempts retry count exceeded")
        log.Printf("[%d] Session failed - %s - %s", requestId, startedGrid.URL.Hostname(), err)
        utils.JsonError(w, err.Error(), http.StatusInternalServerError)
      case context.Canceled:
        log.Printf("[%d] Client disconnected %s - %s - %.2fs", requestId, user, remote, utils.SecondsSince(sessionStartTime))
      }
      startedGrid.Cancel()
      return
    default:
    }
    if err != nil {
      if rsp != nil {
        rsp.Body.Close()
      }
      log.Printf("[%d] Session failed %s", requestId, err)
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
  sessionId := reply["sessionId"].(string)
  log.Printf("[%d] Session Id: %s", requestId, sessionId)
  formattedSessionId, err := utils.GenerateSessionID(sessionId, startedGrid.Name, startedGrid.URL, startedGrid.Grid.Grid.BaseURL, startedGrid.Grid.Grid.VNCPort)
  if err != nil {
    log.Printf("Failed to get formatted session id: %v", err)
    return
  }
  reply["sessionId"] = formattedSessionId
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
    log.Printf("[%d] Location: %v", requestId, l)
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
      s.ID = sessionId
    }
    log.Printf("[%d] Location empty %v", requestId, s.Value.ID)
  }
  if s.ID == "" {
    log.Printf("[%d] Session failed %s", requestId, resp.Status)
    startedGrid.Cancel()
    return
  }

  log.Printf("[%d] Session created with id %s %d in %.2fs", requestId, s.ID, i, utils.SecondsSince(sessionStartTime))
}

// Proxy Handler for all incoming request other than new session
func (h SessionHandler) Proxy(w http.ResponseWriter, r *http.Request) {
  done := make(chan func())
  fragments := strings.Split(r.URL.Path, slash)
  requestId := utils.Serial()
  var name *string
  go func(w http.ResponseWriter, r *http.Request) {
    cancel := func() {}
    defer func() {
      done <- cancel
    }()
    (&httputil.ReverseProxy{
      Director: func(r *http.Request) {

        sessionId, podName, host, port, baseURL, err := utils.ParseSessionID(fragments[2])
        if err != nil {
          log.Printf("[%d] Invalid session id: %v", requestId, err)
        }
        name = &podName
        fragments[2] = sessionId
        r.URL.Path = strings.Join(fragments, slash)
        r.URL.Host, r.URL.Path = host+":"+port, path.Join(baseURL+slash, r.URL.Path)
      },
    }).ServeHTTP(w, r)
    if r.Method == http.MethodDelete && len(fragments) == 3 {
      defer func() {
        err := h.SessionService.Delete(*name)
        if err != nil {
          log.Printf("[%d] Unable to delete pod %s", requestId, name)
        }
      }()
    }
  }(w, r)
  go (<-done)()
}

