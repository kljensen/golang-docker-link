package main

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

func main() {
	ctx := context.Background()

	// Get a Docker API client
	apiClient, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		panic(err)
	}

	// Download the Docker images we'll use
	curlImage := "curlimages/curl"
	httpBinImage := "mccutchen/go-httpbin"
	images := []string{curlImage, httpBinImage}
	for _, image := range images {
		reader, err := apiClient.ImagePull(ctx, image, types.ImagePullOptions{})
		if err != nil {
			panic(err)
		}
		// Log the download
		io.Copy(os.Stdout, reader)
	}

	// This is the "alias" for the httpbin container on
	// our user-defined network.
	httpBinDNSName := "httpbin"

	// Config for httpbin
	httpBinConfig := &container.Config{
		Image: httpBinImage,
		Tty:   false,
	}

	// Config for Curl
	curlConfig := &container.Config{
		Image: "curlimages/curl:latest",
		Cmd:   []string{"curl", "http://" + httpBinDNSName + ":8080/get?foo=bar"},
		Tty:   false,
	}

	// Create a user defined network. Each of our containers
	// will be attached to this network.
	networkID := "kljensen-golang-docker-link"
	networkCreate := types.NetworkCreate{
		Attachable:     true,
		CheckDuplicate: true,
	}
	networkResult, err := apiClient.NetworkCreate(ctx, networkID, networkCreate)
	if err != nil {
		panic(err)
	}

	// Remove the network on panic or when `main` returns
	defer apiClient.NetworkRemove(ctx, networkResult.ID)

	// Create network config for the curl Docker container
	curlNetworkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"net": {
				NetworkID: networkResult.ID,
			},
		},
	}

	// Create Docker config for the httpbin Docker container.
	// Notice how it is given an alias.
	httpBinNetworkConfig := &network.NetworkingConfig{
		EndpointsConfig: map[string]*network.EndpointSettings{
			"net": {
				Aliases:   []string{httpBinDNSName},
				NetworkID: networkResult.ID,
			},
		},
	}

	// Create both containers
	httpBinResp, err := apiClient.ContainerCreate(ctx, httpBinConfig, nil, httpBinNetworkConfig, nil, "")
	if err != nil {
		panic(err)
	}

	if err := apiClient.ContainerStart(ctx, httpBinResp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	curlResp, err := apiClient.ContainerCreate(ctx, curlConfig, nil, curlNetworkConfig, nil, "")
	if err != nil {
		panic(err)
	}

	if err := apiClient.ContainerStart(ctx, curlResp.ID, types.ContainerStartOptions{}); err != nil {
		panic(err)
	}

	// Wait for the curl request to end and its container to stop
	statusCh, errCh := apiClient.ContainerWait(ctx, curlResp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errCh:
		if err != nil {
			panic(err)
		}
	case <-statusCh:
	}

	// Grab the curl logs and echo them to stdout
	out, err := apiClient.ContainerLogs(ctx, curlResp.ID, types.ContainerLogsOptions{ShowStdout: true, ShowStderr: true})
	if err != nil {
		panic(err)
	}
	stdcopy.StdCopy(os.Stdout, os.Stderr, out)

	// Stop the httpbin container
	timeout := 5 * time.Second
	apiClient.ContainerStop(ctx, httpBinResp.ID, &timeout)
}
