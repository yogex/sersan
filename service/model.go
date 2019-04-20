package service

// Caps Browser capabilities
type Caps struct {
    Name             string `json:"browserName"`
    Version          string `json:"version"`
    W3CVersion       string `json:"browserVersion"`
    ScreenResolution string `json:"screenResolution"`
    TestName         string `json:"name"`
    TimeZone         string `json:"timeZone"`
}
