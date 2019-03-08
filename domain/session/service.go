package session

import (
  "github.com/salestock/sersan/config"
  "github.com/salestock/sersan/service"
)

// SessionService Session service
type SessionService struct {
}

// Create Create session
func (s SessionService) Create(browser *Browser, requestId uint64) (service.GridStarter, bool) {
  browserConfig := config.GetBrowserConfig()
  manager := &service.DefaultManager{BrowserConfig: browserConfig}
  return manager.Find(browser.Caps, requestId)
}

// Get Get session
func (s SessionService) Get(podName string) (string, error) {
  conf := config.Get()
  kubernetesClient := service.GetKubernetesClient()
  ip, err := kubernetesClient.WaitUntilReady(podName, conf.StartupTimeout)
  if err != nil {
    return "", err
  }

  return ip, nil
}

// Delete Delete session
func (s SessionService) Delete(podName string) error {
  kubernetesClient := service.GetKubernetesClient()
  err := kubernetesClient.DeletePod(podName)
  if err != nil {
    return err
  }
  return nil
}
