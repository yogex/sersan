package lib

import (
	"errors"
	"fmt"
	"log"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/growbak/hub/config"
	"github.com/growbak/hub/utils"
	apiv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

// KubernetesClient Kubernetes client
type KubernetesClient struct {
	Clientset *kubernetes.Clientset
}

// Kubernetes kubernetes
type Kubernetes struct {
	GridBase GridBase
	Caps     Caps
}

var kubernetesClient *KubernetesClient
var once sync.Once

// GetKubernetesClient Get kubernetes client
func GetKubernetesClient() *KubernetesClient {
	once.Do(func() {
		config, err := rest.InClusterConfig()
		if err != nil {
			log.Printf("Failed to get in cluster config %v", err)
		}
		clientset, err := kubernetes.NewForConfig(config)
		if err != nil {
			log.Printf("Failed to parse config %v", err)
		}
		kubernetesClient = &KubernetesClient{Clientset: clientset}
	})
	return kubernetesClient
}

// CreateGrid Create browsers pod
func (k KubernetesClient) CreateGrid(gridBase *GridBase) (podName string, err error) {
	entryPoint := "/opt/bin/entry_point.sh"
	if gridBase.Grid.EntryPoint != "" {
		entryPoint = gridBase.Grid.EntryPoint
	}
	conf := config.Get()
	podsClient := k.Clientset.CoreV1().Pods(apiv1.NamespaceDefault)
	ports := []apiv1.ContainerPort{
		{
			Name:          "http",
			Protocol:      apiv1.ProtocolTCP,
			ContainerPort: gridBase.Grid.Port,
		},
	}
	if gridBase.Grid.VNCPort != 0 {
		ports = append(ports, apiv1.ContainerPort{
			Name:          "vnc",
			Protocol:      apiv1.ProtocolTCP,
			ContainerPort: gridBase.Grid.VNCPort,
		})
	}

	gridTimeout := conf.GridTimeout
	log.Printf("Caps: %v", gridBase.Timeout)
	if gridBase.Timeout > 0 {
		gridTimeout = gridBase.Timeout
	}
	cpuRequest := conf.CPURequest
	if gridBase.Grid.CPURequest != "" {
		cpuRequest = gridBase.Grid.CPURequest
	}
	cpuLimit := conf.CPULimit
	if gridBase.Grid.CPULimit != "" {
		cpuLimit = gridBase.Grid.CPULimit
	}
	memoryRequest := conf.MemoryRequest
	if gridBase.Grid.MemoryRequest != "" {
		memoryRequest = gridBase.Grid.MemoryRequest
	}
	memoryLimit := conf.MemoryLimit
	if gridBase.Grid.MemoryLimit != "" {
		memoryLimit = gridBase.Grid.MemoryLimit
	}

	spec := &apiv1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			GenerateName: "sersan-grid-" + conf.GridLabel,
			Labels: map[string]string{
				"app": "sersan-grid-" + conf.GridLabel,
			},
		},
		Spec: apiv1.PodSpec{
			Containers: []apiv1.Container{
				{
					Name:    "selenium",
					Image:   gridBase.Grid.Image,
					Ports:   ports,
					Command: []string{"/bin/sh"},
					Args:    []string{"-c", fmt.Sprintf("%s & sleep %d; exit 0", entryPoint, gridTimeout)},
					VolumeMounts: []apiv1.VolumeMount{
						{
							MountPath: "/dev/shm",
							Name:      "dshm",
						},
					},
					Resources: apiv1.ResourceRequirements{
						Limits: apiv1.ResourceList{
							apiv1.ResourceMemory: resource.MustParse(memoryLimit),
							apiv1.ResourceCPU:    resource.MustParse(cpuLimit),
						},
						Requests: apiv1.ResourceList{
							apiv1.ResourceMemory: resource.MustParse(memoryRequest),
							apiv1.ResourceCPU:    resource.MustParse(cpuRequest),
						},
					},
				},
			},
			RestartPolicy: apiv1.RestartPolicyNever,
			Volumes: []apiv1.Volume{
				{
					Name: "dshm",
					VolumeSource: apiv1.VolumeSource{
						EmptyDir: &apiv1.EmptyDirVolumeSource{
							Medium: apiv1.StorageMediumDefault,
						},
					},
				},
			},
		},
	}

	if conf.NodeSelectorKey != "" && conf.NodeSelectorValue != "" {
		spec.Spec.NodeSelector = map[string]string{
			conf.NodeSelectorKey: conf.NodeSelectorValue,
		}
	}

	log.Print("Creating pod")
	pod, err := podsClient.Create(spec)
	if err != nil {
		log.Print("%v", err)
		return
	}
	podName = pod.GetObjectMeta().GetName()
	log.Printf("Pod created - %s", podName)
	return
}

// DeleteGrid Delete pod
func (k KubernetesClient) DeleteGrid(name string) (err error) {
	if !strings.HasPrefix(name, "sersan-grid") {
		err = errors.New("Grid name prefix must be sersan-grid")
		return
	}

	podsClient := k.Clientset.CoreV1().Pods(apiv1.NamespaceDefault)
	err = podsClient.Delete(name, &metav1.DeleteOptions{})
	if err != nil {
		return err
	}

	log.Printf("Pod deleted - %s", name)
	return nil
}

// WaitUntilReady Wait until grid ready
func (k KubernetesClient) WaitUntilReady(name string, timeout int32) (ip string, err error) {
	podsClient := k.Clientset.CoreV1().Pods(apiv1.NamespaceDefault)
	waitTimeout := time.NewTimer(time.Duration(timeout) * time.Millisecond)
	defer waitTimeout.Stop()
	tick := time.Tick(200 * time.Millisecond)
	for {
		select {
		case <-waitTimeout.C:
			err = errors.New(fmt.Sprintf("Pod is not running until %d ms", timeout))
			return
		case <-tick:
			pod, err := podsClient.Get(name, metav1.GetOptions{})
			if err != nil {
				log.Printf("%v", err)
				return "", err
			}
			if pod.Status.Phase == apiv1.PodRunning {
				log.Printf("Pod is ready - %s", name)
				ip = pod.Status.PodIP
				return ip, nil
			}
		}
	}
	return
}

// StartWithCancel Start pod with cancel
func (k Kubernetes) StartWithCancel() (*StartedGrid, error) {
	conf := config.Get()
	kubernetesClient := GetKubernetesClient()
	name, err := kubernetesClient.CreateGrid(&k.GridBase)
	if err != nil {
		return nil, err
	}

	ip, err := kubernetesClient.WaitUntilReady(name, conf.StartupTimeout)
	if err != nil {
		kubernetesClient.DeleteGrid(name)
		return nil, err
	}
	u, err := url.Parse("http://" + ip + ":" + strconv.Itoa(int(k.GridBase.Grid.Port)))
	if err != nil {
		return nil, err
	}

	if k.GridBase.Grid.HealthCheck != "" {
		err = utils.WaitUntilGridReady(u, k.GridBase.Grid.HealthCheck)
		if err != nil {
			kubernetesClient.DeleteGrid(name)
			return nil, err
		}
	}

	s := StartedGrid{
		Name: name,
		URL:  u,
		Grid: k.GridBase,
		Cancel: func() {
			kubernetesClient.DeleteGrid(name)
		},
	}

	return &s, nil
}
