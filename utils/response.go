package utils

import (
    "encoding/json"
    "log"
    "fmt"
    "net/http"

)

// ResponseOk Response OK
func ResponseOk(w http.ResponseWriter, status int, data interface{}) {
    resp := map[string]interface{}{
        "success": "true",
        "data":    data,
        "error":   nil,
    }
    js, err := json.Marshal(resp)
    if err != nil {
        resp := map[string]interface{}{
            "data":  nil,
            "error": fmt.Sprintf("%s", err),
        }
        js, _ = json.Marshal(resp)
    }
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(status)
    w.Write(js)
}

// ResponseFailed Response Failed
func ResponseFailed(w http.ResponseWriter, status int, err error) {
    if status/1e2 == 4 {
        log.Printf("%v", err)
    } else {
        log.Printf("%v", err)
    }

    errMsg := err.Error()
    errResp := map[string]interface{}{
        "code":    "CodeInternalError",
        "message": errMsg,
    }
    resp := map[string]interface{}{
        "success": "false",
        "data":    nil,
        "error":   errResp,
    }
    js, _ := json.Marshal(resp)
    w.Header().Set("Content-Type", "application/json")
    w.WriteHeader(500)
    w.Write(js)
}
