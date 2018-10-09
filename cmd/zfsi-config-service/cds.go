package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/hpidcock/zfsi/pkg/service-agg"
	"github.com/thoas/go-funk"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	config_stream "github.com/hpidcock/zfsi/pkg/config-stream"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

type ClusterDiscoveryService struct {
	cs *config_stream.ConfigStream
}

func (cds *ClusterDiscoveryService) StreamClusters(call envoy.ClusterDiscoveryService_StreamClustersServer) error {
	sub := cds.cs.Register()
	defer sub.Close()

	for {
		req, err := call.Recv()
		if err != nil {
			log.Print(err)
			return err
		}

		config, ok := <-sub.ConfigChan
		if ok == false {
			return status.Errorf(codes.Aborted, "no more config")
		}

		services := config
		if req.ResourceNames != nil {
			resources := funk.Map(req.ResourceNames, func(resourceName string) (string, string) {
				return resourceName, resourceName
			}).(map[string]string)

			services = funk.Filter(config, func(service service_agg.Service) bool {
				_, ok := resources[service.Name]
				return ok
			}).([]service_agg.Service)
		}

		res := &envoy.DiscoveryResponse{
			VersionInfo: time.Now().UTC().String(),
			TypeUrl:     "type.googleapis.com/envoy.api.v2.Cluster",
			Nonce:       req.ResponseNonce,
			Resources:   servicesToClusterConfig(services),
		}

		call.Send(res)
	}
}

func (cds *ClusterDiscoveryService) FetchClusters(ctx context.Context, req *envoy.DiscoveryRequest) (*envoy.DiscoveryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}

func (cds *ClusterDiscoveryService) IncrementalClusters(_ envoy.ClusterDiscoveryService_IncrementalClustersServer) error {
	return errors.New("not implemented")
}

func servicesToClusterConfig(services []service_agg.Service) []types.Any {
	res := make([]types.Any, 0)
	for _, service := range services {

		cluster := &envoy.Cluster{
			Name:                 service.Name,
			Type:                 envoy.Cluster_STRICT_DNS,
			ConnectTimeout:       1000 * time.Millisecond,
			Http2ProtocolOptions: &envoy_core.Http2ProtocolOptions{},
		}

		for _, host := range service.Hosts {
			address := &envoy_core.Address{
				Address: &envoy_core.Address_SocketAddress{
					SocketAddress: &envoy_core.SocketAddress{
						Protocol: envoy_core.TCP,
						Address:  host.Hostname,
						PortSpecifier: &envoy_core.SocketAddress_PortValue{
							PortValue: uint32(host.Port),
						},
					},
				},
			}
			cluster.Hosts = append(cluster.Hosts, address)
		}

		clusterAny, err := types.MarshalAny(cluster)
		if err != nil {
			log.Fatal(err)
		}
		res = append(res, *clusterAny)
	}

	return res
}
