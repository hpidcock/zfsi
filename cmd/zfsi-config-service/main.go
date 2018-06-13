package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/hpidcock/zfsi/pkg/config-stream"
	"github.com/hpidcock/zfsi/pkg/k8s-service-agg"
	"github.com/hpidcock/zfsi/pkg/service-agg"

	"google.golang.org/grpc"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
)

var (
	port                  = flag.Int("port", 53511, "gRPC listen port")
	namespace             = flag.String("namespace", "default", "k8s namespace")
	prefixAnnotation      = flag.String("prefix-annotation", "annotation.zfsi/prefix-list", "annotation on k8s services to use as route prefixes")
	servicePortAnnotation = flag.String("service-port-annotation", "annotation.zfsi/service-port-name", "annotation on k8s services to use to map service ports")
	timeoutAnnotation     = flag.String("timeout-annotation", "annotation.zfsi/timeout", "annotation on k8s services to use for request timeout")
	inCluster             = flag.Bool("in-cluster", true, "use cluster service account or local .kube auth")
)

func main() {
	flag.Parse()

	k8sConfig := k8s_service_agg.Config{
		Namespace:             *namespace,
		PrefixesAnnotation:    *prefixAnnotation,
		ServicePortAnnotation: *servicePortAnnotation,
		TimeoutAnnotation:     *timeoutAnnotation,
		InCluster:             *inCluster,
	}
	serviceAgg, err := k8s_service_agg.NewServiceAggregate(k8sConfig)
	if err != nil {
		log.Fatal(err)
	}

	configStream := config_stream.NewConfigStream()
	go configPoller(configStream, serviceAgg)

	address := fmt.Sprintf(":%d", *port)
	listener, err := net.Listen("tcp", address)

	server := grpc.NewServer()
	cds := &ClusterDiscoveryService{cs: configStream}
	envoy.RegisterClusterDiscoveryServiceServer(server, cds)

	rds := &RouteDiscoveryService{cs: configStream}
	envoy.RegisterRouteDiscoveryServiceServer(server, rds)

	log.Printf("listenting on %s", address)
	err = server.Serve(listener)
	if err != nil {
		log.Fatal(err)
	}
}

func configPoller(configStream *config_stream.ConfigStream, serviceAgg service_agg.ServiceAggregate) {
	ctx := context.Background()
	for {
		config, err := serviceAgg.Services(ctx)
		if err != nil {
			log.Print(err)
		} else {
			configStream.Publish(config)
		}
		time.Sleep(10 * time.Second)
	}
}
