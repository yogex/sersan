package lib

// Caps Browser capabilities
type Caps struct {
	Name                   string `json:"browserName"`
	Version                string `json:"version"`
	W3CVersion             string `json:"browserVersion"`
	ScreenResolution       string `json:"screenResolution"`
	TestName               string `json:"name"`
	TimeZone               string `json:"timeZone"`
	PlatformName           string `json:"platformName"`
	PlatformVersion        string `json:"platformVersion"`
	DeviceName             string `json:"deviceName"`
	App                    string `json:"app"`
	DisableAndroidWatchers string `json:"disableAndroidWatchers"`
	GridTimeout            int    `json:"gridTimeout"`
	NewCommandTimeout      string `json:"newCommandTimeout"`
}
