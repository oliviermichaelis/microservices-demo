// Copyright 2018 Google LLC
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	discountservice "github.com/oliviermichaelis/discount-service/pkg/genproto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"testing"

	pb "github.com/GoogleCloudPlatform/microservices-demo/src/productcatalogservice/genproto"
	"github.com/golang/protobuf/proto"
	"github.com/google/go-cmp/cmp"
	"go.opencensus.io/plugin/ocgrpc"
	"google.golang.org/grpc"
)

type mockedClient struct{}

func (mockedClient) ListProducts(ctx context.Context, in *pb.Empty, opts ...grpc.CallOption) (*pb.ListProductsResponse, error) {
	cat := pb.ListProductsResponse{}
	if err := readCatalogFile(&cat); err != nil {
		return nil, err
	}
	return &cat, nil
}

func (mockedClient) GetProduct(ctx context.Context, in *pb.GetProductRequest, opts ...grpc.CallOption) (*pb.Product, error) {
	panic("implement me")
}

func (mockedClient) SearchProducts(ctx context.Context, in *pb.SearchProductsRequest, opts ...grpc.CallOption) (*pb.SearchProductsResponse, error) {
	panic("implement me")
}

func TestServer(t *testing.T) {
	ctx := context.Background()
	addr := run("0", false)
	conn, err := grpc.Dial(addr,
		grpc.WithInsecure(),
		grpc.WithStatsHandler(&ocgrpc.ClientHandler{}))
	if err != nil {
		t.Fatal(err)
	}
	defer conn.Close()
	client := pb.NewProductCatalogServiceClient(conn)
	res, err := client.ListProducts(ctx, &pb.Empty{})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(res.Products, parseCatalog(), cmp.Comparer(proto.Equal)); diff != "" {
		t.Error(diff)
	}

	got, err := client.GetProduct(ctx, &pb.GetProductRequest{Id: "OLJCESPC7Z"})
	if err != nil {
		t.Fatal(err)
	}
	if want := parseCatalog()[0]; !proto.Equal(got, want) {
		t.Errorf("got %v, want %v", got, want)
	}
	_, err = client.GetProduct(ctx, &pb.GetProductRequest{Id: "N/A"})
	if got, want := status.Code(err), codes.NotFound; got != want {
		t.Errorf("got %s, want %s", got, want)
	}

	sres, err := client.SearchProducts(ctx, &pb.SearchProductsRequest{Query: "typewriter"})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(sres.Results, []*pb.Product{parseCatalog()[0]}, cmp.Comparer(proto.Equal)); diff != "" {
		t.Error(diff)
	}
}

func TestConvertToProduct(t *testing.T) {
	pr := discountservice.Product{
		Id:          "1",
		Name:        "Camera",
		Description: "vintage camera",
		Picture:     "123",
		PriceUsd:    &discountservice.Money{},
		Categories:  []string{"hobbies", "vintage"},
		Discount:    25,
	}

	prc := convertToProduct(&pr)
	if pr.Id != prc.Id {
		t.Errorf("Id should be %s but is %s", pr.Id, prc.Id)
	}
	if pr.Name != prc.Name {
		t.Errorf("Name should be %s but is %s", pr.Name, prc.Name)
	}
	if pr.Description != prc.Description {
		t.Errorf("Description should be %s but is %s", pr.Description, prc.Description)
	}
	if pr.Picture != prc.Picture {
		t.Errorf("Picture should be %s but is %s", pr.Picture, prc.Picture)
	}
	if pr.Discount != prc.Discount {
		t.Errorf("Id should be %d but is %d", pr.Discount, prc.Discount)
	}
}

func TestConvertToProductDiscount(t *testing.T) {
	pr := pb.Product{
		Id:          "1",
		Name:        "Camera",
		Description: "vintage camera",
		Picture:     "123",
		PriceUsd:    &pb.Money{},
		Categories:  []string{"hobbies", "vintage"},
		Discount:    25,
	}

	prc := convertToProductDiscount(&pr)
	if pr.Id != prc.Id {
		t.Errorf("Id should be %s but is %s", pr.Id, prc.Id)
	}
	if pr.Name != prc.Name {
		t.Errorf("Name should be %s but is %s", pr.Name, prc.Name)
	}
	if pr.Description != prc.Description {
		t.Errorf("Description should be %s but is %s", pr.Description, prc.Description)
	}
	if pr.Picture != prc.Picture {
		t.Errorf("Picture should be %s but is %s", pr.Picture, prc.Picture)
	}
	if pr.Discount != prc.Discount {
		t.Errorf("Id should be %d but is %d", pr.Discount, prc.Discount)
	}
}
