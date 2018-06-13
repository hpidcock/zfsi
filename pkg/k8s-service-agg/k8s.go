package k8s_service_agg

import (
	"context"
	"errors"
	"log"
	"os/user"
	"path"
	"strings"
	"time"

	"github.com/thoas/go-funk"

	service_agg "github.com/hpidcock/zfsi/pkg/service-agg"

	core_v1 "k8s.io/api/core/v1"
	meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"

	// Handle GKE auth
	_ "k8s.io/client-go/plugin/pkg/client/auth/gcp"
)

// ServiceAggregate provides k8s services
type ServiceAggregate struct {
	config    Config
	clientset *kubernetes.Clientset
}

var (
	errInvalidService = errors.New("invalid")
)

// NewServiceAggregate returns a new k8s ServiceAggregate
func NewServiceAggregate(config Config) (*ServiceAggregate, error) {
	var err error
	sa := &ServiceAggregate{
		config: config,
	}

	var k8sConfig *rest.Config
	if config.InCluster {
		k8sConfig, err = rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
	} else {
		usr, err := user.Current()
		if err != nil {
			return nil, err
		}

		homeDir := path.Join(usr.HomeDir, ".kube/config")
		k8sConfig, err = clientcmd.BuildConfigFromFlags("", homeDir)
		if err != nil {
			return nil, err
		}
	}

	sa.clientset, err = kubernetes.NewForConfig(k8sConfig)
	if err != nil {
		return nil, err
	}

	return sa, nil
}

// Services implements service_agg.ServiceAggregate.Services
func (sa *ServiceAggregate) Services(context.Context) ([]service_agg.Service, error) {
	res, err := sa.clientset.CoreV1().
		Services(sa.config.Namespace).
		List(meta_v1.ListOptions{})
	if err != nil {
		return nil, err
	}

	services := make([]service_agg.Service, 0)
	for _, service := range res.Items {
		svc, err := sa.parseService(service)
		if err == errInvalidService {
			continue
		} else if err != nil {
			return nil, err
		}

		services = append(services, svc)
	}

	return services, nil
}

func (sa *ServiceAggregate) parseService(ksvc core_v1.Service) (service_agg.Service, error) {
	var err error
	if ksvc.Spec.Type != core_v1.ServiceTypeClusterIP &&
		ksvc.Spec.Type != core_v1.ServiceTypeNodePort &&
		ksvc.Spec.Type != core_v1.ServiceTypeLoadBalancer {
		return service_agg.Service{}, errInvalidService
	}

	prefixString, ok := ksvc.ObjectMeta.Annotations[sa.config.PrefixesAnnotation]
	if ok == false {
		return service_agg.Service{}, errInvalidService
	}

	portName, ok := ksvc.ObjectMeta.Annotations[sa.config.ServicePortAnnotation]
	if ok == false {
		return service_agg.Service{}, errInvalidService
	}

	timeoutString, ok := ksvc.ObjectMeta.Annotations[sa.config.TimeoutAnnotation]
	timeout := 15 * time.Second // Envoy default
	if ok {
		timeout, err = time.ParseDuration(timeoutString)
		if err != nil {
			return service_agg.Service{}, err
		}
	}

	ports := funk.Filter(ksvc.Spec.Ports, func(port core_v1.ServicePort) bool {
		return port.Name == portName &&
			port.Protocol == core_v1.ProtocolTCP
	}).([]core_v1.ServicePort)

	if len(ports) != 1 {
		log.Printf("service %s: unable to match port %s", ksvc.Name, portName)
		return service_agg.Service{}, errInvalidService
	}

	svc := service_agg.Service{}
	svc.Name = ksvc.Name
	svc.Prefixes = strings.Split(prefixString, ",")
	svc.Hostname = ksvc.Spec.ClusterIP
	svc.Port = int(ports[0].Port)
	svc.Timeout = timeout
	return svc, nil
}
