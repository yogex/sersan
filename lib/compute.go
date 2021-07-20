package lib

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/salestock/sersan/config"
	"github.com/salestock/sersan/utils"

	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	compute "google.golang.org/api/compute/v1"
)

type ComputeClient struct {
	Clientset *http.Client
}

type ComputeEngine struct {
	GridBase GridBase
	Caps     Caps
}

var computeClient ComputeClient
var computeOnce sync.Once

func GetComputeClient() ComputeClient {
	computeOnce.Do(func() {
		client, err := google.DefaultClient(oauth2.NoContext, compute.ComputeScope)
		if err != nil {
			log.Printf("Failed to get compute client: %v", err)
		}
		computeClient = ComputeClient{Clientset: client}
	})
	return computeClient
}

func (c ComputeClient) CreateGrid(gridBase *GridBase) (name string, err error) {
	conf := config.Get()
	service, err := compute.New(c.Clientset)
	if err != nil {
		log.Printf("Failed to get service: %v", err)
	}

	prefix := "https://www.googleapis.com/compute/v1/projects/" + conf.ProjectID
	startupScript := "gs://" + conf.BucketName + "/startup.sh"
	computeName := "sersan-grid-" + conf.GridLabel + "-" + utils.GenerateUUID()
	machineType := conf.MachineType
	if gridBase.Grid.MachineType != "" {
		machineType = gridBase.Grid.MachineType
	}
	instance := &compute.Instance{
		Name:        computeName,
		Description: "Android Emulator Runner",
		MachineType: prefix + "/zones/" + conf.Zone + "/machineTypes/" + machineType,
		Disks: []*compute.AttachedDisk{
			{
				AutoDelete: true,
				Boot:       true,
				Type:       "PERSISTENT",
				InitializeParams: &compute.AttachedDiskInitializeParams{
					DiskName:    computeName,
					SourceImage: gridBase.Grid.Image,
				},
			},
		},
		NetworkInterfaces: []*compute.NetworkInterface{
			{
				AccessConfigs: []*compute.AccessConfig{
					{
						Type: "ONE_TO_ONE_NAT",
						Name: "External NAT",
					},
				},
				Subnetwork: conf.Subnetwork,
			},
		},
		Scheduling: &compute.Scheduling{
			Preemptible: true,
		},
		ServiceAccounts: []*compute.ServiceAccount{
			{
				Email: "default",
				Scopes: []string{
					compute.DevstorageFullControlScope,
					compute.ComputeScope,
				},
			},
		},
		Metadata: &compute.Metadata{
			Items: []*compute.MetadataItems{
				{
					Key:   "startup-script-url",
					Value: &startupScript,
				},
			},
		},
		Labels: map[string]string{
			"service-name": "sersan-grid",
		},
		Tags: &compute.Tags{
			Items: []string{"vnc-server", "appium"},
		},
	}

	log.Printf("%v", instance)
	_, err = service.Instances.Insert(conf.ProjectID, conf.Zone, instance).Do()
	if err != nil {
		log.Printf("%v", err)
		return computeName, err
	}

	return computeName, nil
}

func (c ComputeClient) DeleteGrid(name string) (err error) {
	if !strings.HasPrefix(name, "sersan-grid") {
		err = errors.New("Grid name prefix must be sersan-grid")
		return
	}

	conf := config.Get()
	service, err := compute.New(c.Clientset)
	if err != nil {
		return err
	}

	_, err = service.Instances.Delete(conf.ProjectID, conf.Zone, name).Do()
	if err != nil {
		return err
	}

	return
}

func (c ComputeClient) WaitUntilReady(name string, timeout int32) (ip string, err error) {
	conf := config.Get()
	service, err := compute.New(c.Clientset)
	if err != nil {
		return
	}
	waitTimeout := time.NewTimer(time.Duration(timeout) * time.Millisecond)
	defer waitTimeout.Stop()
	tick := time.Tick(200 * time.Millisecond)
	for {
		select {
		case <-waitTimeout.C:
			err = fmt.Errorf("Grid is not running until %d ms", timeout)
			return
		case <-tick:
			instance, err := service.Instances.Get(conf.ProjectID, conf.Zone, name).Do()
			if err != nil {
				log.Printf("Failed to get instance: %v", err)
			}

			if instance.Status == "RUNNING" {
				if conf.ExternalIP {
					if len(instance.NetworkInterfaces[0].AccessConfigs) > 0 {
						return instance.NetworkInterfaces[0].AccessConfigs[0].NatIP, nil
					}
					err = fmt.Errorf("External IP not found")
					return "", err
				}

				if instance.NetworkInterfaces[0].NetworkIP != "" {
					return instance.NetworkInterfaces[0].NetworkIP, nil
				}
				err = fmt.Errorf("External IP not found")
				return "", err
			}
		}
	}
	return
}

func (ce ComputeEngine) StartWithCancel() (grid *StartedGrid, err error) {
	conf := config.Get()
	computeClient := GetComputeClient()
	name, err := computeClient.CreateGrid(&ce.GridBase)
	if err != nil {
		return nil, err
	}

	ip, err := computeClient.WaitUntilReady(name, conf.StartupTimeout)
	if err != nil {
		computeClient.DeleteGrid(name)
		return nil, err
	}
	u, err := url.Parse("http://" + ip + ":" + strconv.Itoa(int(ce.GridBase.Grid.Port)))
	if err != nil {
		return nil, err
	}

	if ce.GridBase.Grid.HealthCheck != "" {
		err = utils.WaitUntilGridReady(u, ce.GridBase.Grid.HealthCheck)
		if err != nil {
			computeClient.DeleteGrid(name)
			return nil, err
		}
	}

	s := StartedGrid{
		Name: name,
		URL:  u,
		Grid: ce.GridBase,
		Cancel: func() {
			kubernetesClient.DeleteGrid(name)
		},
	}

	return &s, nil
}
