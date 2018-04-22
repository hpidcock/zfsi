package main

import (
	"context"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/thoas/go-funk"

	config_stream "github.com/hpidcock/zfsi/pkg/config-stream"

	envoy "github.com/envoyproxy/go-control-plane/envoy/api/v2"
	envoy_core "github.com/envoyproxy/go-control-plane/envoy/api/v2/core"
	envoy_route "github.com/envoyproxy/go-control-plane/envoy/api/v2/route"
	codes "google.golang.org/grpc/codes"
	status "google.golang.org/grpc/status"
)

const (
	routeConfigName = "zfsi-route-config"
)

type RouteDiscoveryService struct {
	cs *config_stream.ConfigStream
}

func (rds *RouteDiscoveryService) StreamRoutes(call envoy.RouteDiscoveryService_StreamRoutesServer) error {
	sub := rds.cs.Register()
	defer sub.Close()

	for {
		req, err := call.Recv()
		if err != nil {
			return err
		}

		config, ok := <-sub.ConfigChan
		if ok == false {
			return status.Errorf(codes.Aborted, "no more config")
		}

		res := &envoy.DiscoveryResponse{
			VersionInfo: time.Now().UTC().String(),
			TypeUrl:     "type.googleapis.com/envoy.api.v2.RouteConfiguration",
			Nonce:       req.ResponseNonce,
			Resources:   make([]types.Any, 0),
		}

		if funk.ContainsString(req.ResourceNames, routeConfigName) == false {
			call.Send(res)
			continue
		}

		routes := make([]envoy_route.Route, 0)
		healthRoute := envoy_route.Route{
			Match: envoy_route.RouteMatch{
				PathSpecifier: &envoy_route.RouteMatch_Prefix{
					Prefix: "/healthz",
				},
			},
			Action: &envoy_route.Route_DirectResponse{
				DirectResponse: &envoy_route.DirectResponseAction{
					Status: 200,
					Body: &envoy_core.DataSource{
						Specifier: &envoy_core.DataSource_InlineString{
							InlineString: "OK",
						},
					},
				},
			},
		}
		routes = append(routes, healthRoute)

		for _, service := range config {
			for _, prefix := range service.Prefixes {
				route := envoy_route.Route{
					Match: envoy_route.RouteMatch{
						PathSpecifier: &envoy_route.RouteMatch_Prefix{
							Prefix: "/" + prefix,
						},
					},
					Action: &envoy_route.Route_Route{
						Route: &envoy_route.RouteAction{
							ClusterSpecifier: &envoy_route.RouteAction_Cluster{
								Cluster: service.Name,
							},
						},
					},
				}

				routes = append(routes, route)
			}
		}

		virtualHost := envoy_route.VirtualHost{
			Name:    "wildcard",
			Domains: []string{"*"},
			Routes:  routes,
			Cors: &envoy_route.CorsPolicy{
				Enabled: &types.BoolValue{
					Value: true,
				},
				AllowOrigin:  []string{"*"},
				AllowHeaders: "authorization,content-type,x-grpc-web",
				AllowMethods: "POST",
				AllowCredentials: &types.BoolValue{
					Value: true,
				},
			},
		}

		routeConfig := &envoy.RouteConfiguration{
			Name:         routeConfigName,
			VirtualHosts: []envoy_route.VirtualHost{virtualHost},
		}

		routeConfigAny, err := types.MarshalAny(routeConfig)
		if err != nil {
			return err
		}
		res.Resources = append(res.Resources, *routeConfigAny)

		call.Send(res)
	}
}

func (rds *RouteDiscoveryService) FetchRoutes(ctx context.Context, req *envoy.DiscoveryRequest) (*envoy.DiscoveryResponse, error) {
	return nil, status.Errorf(codes.Unimplemented, "unimplemented")
}
