package service

import (
    "errors"
    "fmt"
    "log"
    "net/http"
    "net/url"
    "strconv"
    "sync"
    "time"

    "github.com/salestock/sersan/config"
    apiv1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/resource"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
)

const (
    NamespaceSersan = "sersan"
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

// WaitUntilSeleniumReady Wait until selenium ready
func WaitUntilSeleniumReady(url *url.URL, healthCheck string) (err error) {
    conf := config.Get()
    waitTimeout := time.NewTimer(time.Duration(conf.SeleniumStartupTimeout) * time.Millisecond)
    defer waitTimeout.Stop()
    tick := time.Tick(200 * time.Millisecond)
    for {
        select {
        case <-waitTimeout.C:
            err = errors.New(fmt.Sprintf("Timeout - Selenium is not ready until %v", waitTimeout))
            return
        case <-tick:
            resp, _ := http.Get(url.String() + healthCheck)

            if resp != nil {
                defer resp.Body.Close()
                if resp.StatusCode == 200 {
                    log.Print("Selenium is ready")
                    return nil
                }
            }
        }
    }
    return
}

// CreatePod Create pod
func (k KubernetesClient) CreatePod(gridBase *GridBase) (podName string, err error) {
    entryPoint := "/opt/bin/entry_point.sh"
    if gridBase.Grid.EntryPoint != "" {
        entryPoint = gridBase.Grid.EntryPoint
    }
    conf := config.Get()
    podsClient := k.Clientset.CoreV1().Pods(NamespaceSersan)
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
                    Name:  "selenium",
                    Image: gridBase.Grid.Image,
                    Ports: ports,
                    Command: []string { "/bin/sh" },
                    Args: []string {"-c", fmt.Sprintf("%s & sleep %d; exit 0", entryPoint, conf.GridTimeout)},
                    VolumeMounts: []apiv1.VolumeMount{
                        {
                            MountPath: "/dev/shm",
                            Name:      "dshm",
                        },
                    },
                    Resources: apiv1.ResourceRequirements{
                        Limits: apiv1.ResourceList{
                            apiv1.ResourceMemory: resource.MustParse(conf.MemoryLimit),
                            apiv1.ResourceCPU:    resource.MustParse(conf.CPULimit),
                        },
                        Requests: apiv1.ResourceList{
                            apiv1.ResourceMemory: resource.MustParse(conf.MemoryRequest),
                            apiv1.ResourceCPU:    resource.MustParse(conf.CPURequest),
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

// WaitUntilReady Wait until pod ready
func (k KubernetesClient) WaitUntilReady(name string, timeout int32) (ip string, err error) {
    podsClient := k.Clientset.CoreV1().Pods(NamespaceSersan)
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

// DeletePod Delete pod
func (k KubernetesClient) DeletePod(name string) error {
    podsClient := k.Clientset.CoreV1().Pods(NamespaceSersan)
    err := podsClient.Delete(name, &metav1.DeleteOptions{})
    if err != nil {
        return err
    }

    log.Printf("Pod deleted - %s", name)
    return nil
}

// StartWithCancel Start pod with cancel
func (k *Kubernetes) StartWithCancel() (*StartedGrid, error) {
    conf := config.Get()
    kubernetesClient := GetKubernetesClient()
    name, err := kubernetesClient.CreatePod(&k.GridBase)
    if err != nil {
        return nil, err
    }

    ip, err := kubernetesClient.WaitUntilReady(name, conf.StartupTimeout)
    if err != nil {
        kubernetesClient.DeletePod(name)
        return nil, err
    }
    u, err := url.Parse("http://" + ip + ":" + strconv.Itoa(int(k.GridBase.Grid.Port)))
    if err != nil {
        return nil, err
    }

    if k.GridBase.Grid.HealthCheck != "" {
        err = WaitUntilSeleniumReady(u, k.GridBase.Grid.HealthCheck)
        if err != nil {
            kubernetesClient.DeletePod(name)
            return nil, err
        }
    }

    s := StartedGrid{
        Name: name,
        URL:  u,
        Grid: k.GridBase,
        Cancel: func() {
            kubernetesClient.DeletePod(name)
        },
    }

    return &s, nil
}
