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
	"errors"
	"fmt"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

// DockerService contains all settings to talk to the docker api
type DockerService struct {
	// the context docker is relate to
	Context context.Context
	// client is the interface to the docker runtime
	Client *client.Client
	Cancel context.CancelFunc
}

func NewDockerServiceWithTimeout(timeout time.Duration) *DockerService {
	// ctx := context.Background()
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	// defer cancel()
	cli, err := client.NewEnvClient()
	if err != nil {
		panic(err)
	}

	return &DockerService{
		Context: ctx,
		Client:  cli,
		Cancel:  cancel,
	}
}

// ListContainers lists all docker containers
func (ds *DockerService) ListContainers() {

	containers, err := ds.Client.ContainerList(ds.Context, types.ContainerListOptions{})
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

	images, err := ds.Client.ImageList(ds.Context, types.ImageListOptions{})
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
	// image example: "docker.io/library/alpine"
	outputReader, err := ds.Client.ImagePull(ds.Context, image, types.ImagePullOptions{})
	if err != nil {
		return "", err
	}
	// io.Copy(os.Stdout, reader)

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
	// create some context for docker

	// fmt.Println("imageName", imageName)
	// fmt.Println("submissionZipFile", submissionZipFile)
	// fmt.Println("frameworkZipFile", frameworkZipFile)

	// submissionZipFile := "/home/patwie/git/github.com/infomark-org/infomark/infomark-backend/.local/simple_ci_runner/submission.zip"
	// frameworkZipFile := "/home/patwie/git/github.com/infomark-org/infomark/infomark-backend/.local/simple_ci_runner/unittest.zip"
	defer ds.Cancel()
	cmds := []string{}

	cfg := &container.Config{
		Image:           imageName,
		Cmd:             cmds,
		Tty:             true,
		AttachStdin:     false,
		AttachStdout:    true,
		AttachStderr:    true,
		NetworkDisabled: true, // no network activity required
		// StopTimeout:
	}

	hostCfg := &container.HostConfig{
		Resources: container.Resources{
			Memory:     DockerMemoryBytes, // 200mb
			MemorySwap: 0,
		},
		// AutoRemove: true,
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

	resp, err := ds.Client.ContainerCreate(ds.Context, cfg, hostCfg, nil, "")
	if err != nil {
		return "", 0, err
	}

	defer ds.Client.ContainerRemove(ds.Context, resp.ID, types.ContainerRemoveOptions{})

	if err := ds.Client.ContainerStart(ds.Context, resp.ID, types.ContainerStartOptions{}); err != nil {
		return "", 0, err
	}

	_, errC := ds.Client.ContainerWait(ds.Context, resp.ID, "")

	err = <-errC
	if err != nil {
		if errors.Is(err, context.DeadlineExceeded) {
			return "Execution took to long", 0, nil

		}
		return err.Error(), 0, err
	}

	outputReader, err := ds.Client.ContainerLogs(ds.Context, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return "", 0, err
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(outputReader)

	// io.Copy(os.Stdout, outputReader)
	return buf.String(), 0, nil
}
