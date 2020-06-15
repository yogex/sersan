package session

import (
	"github.com/salestock/sersan/lib"
)

// SessionService Session service
type SessionService struct {
}

// Create Create session
func (s SessionService) Create(browser *Browser) (lib.GridStarter, bool) {
	gridConfig := lib.GetGridConfig()
	manager := &lib.DefaultManager{GridConfig: gridConfig}
	return manager.Find(browser.Caps)
}

// Delete Delete session
func (s SessionService) Delete(name string, engine string) error {
	client := lib.GetEngineClient(engine)
	err := client.DeleteGrid(name)
	if err != nil {
		return err
	}
	return nil
}
