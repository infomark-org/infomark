// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019 ComputerGraphics Tuebingen
//               2020-present InfoMark.org
// Authors: Patrick Wieschollek
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.

// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.

// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package service

import (
	"bytes"
	"context"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

// DockerService contains all settings to talk to the docker api
type DockerService struct {
	Client  *client.Client
	Timeout time.Duration
}

func NewDockerServiceWithTimeout(timeout time.Duration) (*DockerService, error) {
	cli, err := client.NewClientWithOpts(client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}

	return &DockerService{
		Timeout: timeout,
		Client:  cli,
	}, nil
}

// ListContainers lists all docker containers
func (ds *DockerService) ListContainers() {
	ctx := context.Background()
	containers, err := ds.Client.ContainerList(ctx, types.ContainerListOptions{})
	if err != nil {
		panic(err)
	}

	for _, container := range containers {
		fmt.Println(container.ID)
		fmt.Println(container.Names)
	}
}

// ListImages lists all docker images
func (ds *DockerService) ListImages() {
	ctx := context.Background()

	images, err := ds.Client.ImageList(ctx, types.ImageListOptions{})
	if err != nil {
		panic(err)
	}

	for _, image := range images {
		fmt.Println(image.ID)
		fmt.Println(image.RepoTags)
		fmt.Println(image.Size)
		fmt.Println(image.VirtualSize)
		if len(image.RepoTags) > 0 {
			fmt.Println(image.RepoTags[0])
		}
		fmt.Println("")
	}
}

// Pull pulls a docker image
func (ds *DockerService) Pull(image string) (string, error) {
	ctx := context.Background()
	// image example: "docker.io/library/alpine"
	outputReader, err := ds.Client.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return "", err
	}
	buf := new(bytes.Buffer)
	buf.ReadFrom(outputReader)

	return buf.String(), nil

}

// Run executes a docker container and waits for the output
func (ds *DockerService) Run(
	imageName string,
	submissionZipFile string,
	frameworkZipFile string,
	DockerMemoryBytes int64,
) (string, int64, error) {
	ctx, cancel := context.WithTimeout(context.Background(), ds.Timeout)
	defer cancel()
	cmds := []string{}

	cfg := &container.Config{
		Image:           imageName,
		Cmd:             cmds,
		Tty:             true,
		AttachStdin:     false,
		AttachStdout:    true,
		AttachStderr:    true,
		NetworkDisabled: true, // no network activity required
	}

	// See https://docs.docker.com/config/containers/resource_constraints/#cpu
	// Each Worker gets something equivalent to 1 core. If you have 4 cores, this
	// will allow each worker to get 100% (eg. 25% per core).
	cpu_maximum := int64(100000)

	hostCfg := &container.HostConfig{
		Resources: container.Resources{
			CPUPeriod:  cpu_maximum,
			CPUQuota:   cpu_maximum,
			Memory:     DockerMemoryBytes,
			MemorySwap: 0,
		},
		Mounts: []mount.Mount{
			{
				ReadOnly: true,
				Type:     mount.TypeBind,
				Source:   submissionZipFile,
				Target:   "/data/submission.zip",
			},
			{
				ReadOnly: true,
				Type:     mount.TypeBind,
				Source:   frameworkZipFile,
				Target:   "/data/unittest.zip",
			},
		},
	}

	resp, err := ds.Client.ContainerCreate(ctx, cfg, hostCfg, nil, nil, "")
	if err != nil {
		return "", 0, err
	}

	defer ds.Client.ContainerRemove(ctx, resp.ID, types.ContainerRemoveOptions{Force:true})

	if err := ds.Client.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", 0, err
	}

	defer ds.Client.ContainerKill(context.Background(), resp.ID, "9")

	statusCh, errCh := ds.Client.ContainerWait(ctx, resp.ID, "")
	select {
	case <-ctx.Done():
		return "Execution took too long (Timeout: "+ds.Timeout.String()+")", 0, nil
	case err := <-errCh:
		return err.Error(), 0, err
	case <-statusCh:
	}

	outputReader, err := ds.Client.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return "", 0, err
	}

	buf := new(bytes.Buffer)
	len, err := buf.ReadFrom(outputReader)

	// avoid submitting large outputs to the database
	// postgres will not accept more than 64kB and we don't want that much either
	if (len > 32*1024) {
		return "Output too large (you're printing too much stuff)", 0, nil
	}

	return buf.String(), 0, nil
}
