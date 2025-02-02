// Copyright 2022 TiKV Project Authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"context"
	"net/http"

	"github.com/pingcap/errors"
	rmpb "github.com/pingcap/kvproto/pkg/resource_manager"
	"github.com/tikv/pd/pkg/mcs/registry"
	"github.com/tikv/pd/server"
	"google.golang.org/grpc"
)

var _ rmpb.ResourceManagerServer = (*Service)(nil)

// SetUpRestHandler is a hook to sets up the REST service.
var SetUpRestHandler = func(srv *Service) (http.Handler, server.APIServiceGroup) {
	return dummyRestService{}, server.APIServiceGroup{}
}

type dummyRestService struct{}

func (d dummyRestService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNotImplemented)
	w.Write([]byte("not implemented"))
}

// Service is the gRPC service for resource manager.
type Service struct {
	ctx     context.Context
	manager *Manager
	// settings
}

// NewService creates a new resource manager service.
func NewService(svr *server.Server) registry.RegistrableService {
	manager := NewManager(svr)

	return &Service{
		ctx:     svr.Context(),
		manager: manager,
	}
}

// RegisterGRPCService registers the service to gRPC server.
func (s *Service) RegisterGRPCService(g *grpc.Server) {
	rmpb.RegisterResourceManagerServer(g, s)
}

// RegisterRESTHandler registers the service to REST server.
func (s *Service) RegisterRESTHandler(userDefineHandlers map[string]http.Handler) {
	handler, group := SetUpRestHandler(s)
	server.RegisterUserDefinedHandlers(userDefineHandlers, &group, handler)
}

// GetManager returns the resource manager.
func (s *Service) GetManager() *Manager {
	return s.manager
}

// GetResourceGroup implements ResourceManagerServer.GetResourceGroup.
func (s *Service) GetResourceGroup(ctx context.Context, req *rmpb.GetResourceGroupRequest) (*rmpb.GetResourceGroupResponse, error) {
	rg := s.manager.GetResourceGroup(req.ResourceGroupName)
	if rg == nil {
		return nil, errors.New("resource group not found")
	}
	return &rmpb.GetResourceGroupResponse{
		Group: rg.IntoProtoResourceGroup(),
	}, nil
}

// ListResourceGroups implements ResourceManagerServer.ListResourceGroups.
func (s *Service) ListResourceGroups(ctx context.Context, req *rmpb.ListResourceGroupsRequest) (*rmpb.ListResourceGroupsResponse, error) {
	groups := s.manager.GetResourceGroupList()
	resp := &rmpb.ListResourceGroupsResponse{
		Groups: make([]*rmpb.ResourceGroup, 0, len(groups)),
	}
	for _, group := range groups {
		resp.Groups = append(resp.Groups, group.IntoProtoResourceGroup())
	}
	return resp, nil
}

// AddResourceGroup implements ResourceManagerServer.AddResourceGroup.
func (s *Service) AddResourceGroup(ctx context.Context, req *rmpb.PutResourceGroupRequest) (*rmpb.PutResourceGroupResponse, error) {
	rg := FromProtoResourceGroup(req.GetGroup())
	err := s.manager.AddResourceGroup(rg)
	if err != nil {
		return nil, err
	}
	return &rmpb.PutResourceGroupResponse{Body: "Success!"}, nil
}

// DeleteResourceGroup implements ResourceManagerServer.DeleteResourceGroup.
func (s *Service) DeleteResourceGroup(ctx context.Context, req *rmpb.DeleteResourceGroupRequest) (*rmpb.DeleteResourceGroupResponse, error) {
	err := s.manager.DeleteResourceGroup(req.ResourceGroupName)
	if err != nil {
		return nil, err
	}
	return &rmpb.DeleteResourceGroupResponse{Body: "Success!"}, nil
}

// ModifyResourceGroup implements ResourceManagerServer.ModifyResourceGroup.
func (s *Service) ModifyResourceGroup(ctx context.Context, req *rmpb.PutResourceGroupRequest) (*rmpb.PutResourceGroupResponse, error) {
	err := s.manager.ModifyResourceGroup(req.GetGroup())
	if err != nil {
		return nil, err
	}
	return &rmpb.PutResourceGroupResponse{Body: "Success!"}, nil
}

// AcquireTokenBuckets implements ResourceManagerServer.AcquireTokenBuckets.
func (s *Service) AcquireTokenBuckets(stream rmpb.ResourceManager_AcquireTokenBucketsServer) error {
	return errors.New("Not implemented")
}
