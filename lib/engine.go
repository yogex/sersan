package lib

import (
	"log"
	"strings"
)

const (
	KubernetesType    = "kubernetes"
	ComputeEngineType = "compute"
)

type Engine interface {
	CreateGrid(gridBase *GridBase) (string, error)
	DeleteGrid(name string) error
	WaitUntilReady(name string, timeout int32) (string, error)
}

func GetEngineClient(engineType string) (engine Engine) {
	switch strings.ToLower(engineType) {
	case KubernetesType:
		log.Printf("Get kubernetes client")
		return GetKubernetesClient()
	case ComputeEngineType:
		log.Printf("Get compute engine client")
		return GetComputeClient()
	default:
		log.Printf("Get default kubernetes client")
		return GetKubernetesClient()
	}
}

func GetGridStarter(engineType string, gridBase GridBase, caps Caps) (grid GridStarter) {
	switch strings.ToLower(engineType) {
	case KubernetesType:
		log.Printf("Get kubernetes")
		return Kubernetes{
			GridBase: gridBase,
			Caps:     caps,
		}
	case ComputeEngineType:
		log.Printf("Get compute engine")
		return ComputeEngine{
			GridBase: gridBase,
			Caps:     caps,
		}
	default:
		log.Printf("Get default engine")
		return Kubernetes{
			GridBase: gridBase,
			Caps:     caps,
		}
	}
}
