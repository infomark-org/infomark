// InfoMark - a platform for managing courses with
//            distributing exercise sheets and testing exercise submissions
// Copyright (C) 2019  ComputerGraphics Tuebingen
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

package main

// // SubmissionPayload holds all information necessary to pull the file, run the tests
// // and publish the result
// type SubmissionPayload struct {
//   SubmissionID int
//   Dockerfile   string
// }

// type DockerService struct {
//   // the context docker is relate to
//   Context context.Context
//   // client is the interface to the docker runtime
//   Client *client.Client
// }

// func NewDockerService() *DockerService {
//     ctx := context.Background()
//   cli, err := client.NewEnvClient()
//   if err != nil {
//     panic(err)
//   }

//   return &DockerService{
//     Context: ctx,
//     Client:  cli,
//   }
// }

// func (ds *DockerService) ListContainers() {

//   containers, err := ds.Client.ContainerList(ds.Context, types.ContainerListOptions{})
//   if err != nil {
//     panic(err)
//   }

//   for _, container := range containers {
//     fmt.Println(container.ID)
//     fmt.Println(container.Names)
//   }
// }

// func (ds *DockerService) ListImages() {

//   images, err := ds.Client.ImageList(ds.Context, types.ImageListOptions{})
//   if err != nil {
//     panic(err)
//   }

//   for _, image := range images {
//     fmt.Println(image.ID)
//     fmt.Println(image.RepoTags)
//     fmt.Println(image.Size)
//     fmt.Println(image.VirtualSize)
//     if len(image.RepoTags) > 0 {
//       fmt.Println(image.RepoTags[0])
//     }
//     fmt.Println("")
//   }
// }

// func (ds *DockerService) Pull(image string) (string, error) {
//   // image example: "docker.io/library/alpine"
//   outputReader, err := ds.Client.ImagePull(ds.Context, image, types.ImagePullOptions{})
//   if err != nil {
//     return "", err
//   }
//   // io.Copy(os.Stdout, reader)

//   buf := new(bytes.Buffer)
//   buf.ReadFrom(outputReader)

//   return buf.String(), nil

// }

// // Run executes a docker conter and waits for the output
// func (ds *DockerService) Run(image string, cmd []string) (string, int64, error) {
//   // create some context for docker

//   cfg := &container.Config{
//     Image: image,
//     Cmd:   cmd,
//     Tty:   true,
//   }

//   resp, err := ds.Client.ContainerCreate(ds.Context, cfg, nil, nil, "")
//   if err != nil {
//     return "", 0, err
//   }

//   if err := ds.Client.ContainerStart(ds.Context, resp.ID, types.ContainerStartOptions{}); err != nil {
//     return "", 0, err
//   }

//   exitCode, err := ds.Client.ContainerWait(ds.Context, resp.ID)
//   outputReader, err := ds.Client.ContainerLogs(ds.Context, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
//   if err != nil {
//     return "", 0, err
//   }

//   buf := new(bytes.Buffer)
//   buf.ReadFrom(outputReader)

//   // io.Copy(os.Stdout, outputReader)
//   return buf.String(), exitCode, nil
// }

// func main() {

//   ds := NewDockerService()

//   ds.ListImages()

//   output, exitCode, err := ds.Run("alpine", []string{"echo", "hello world"})
//   fmt.Println(output)
//   fmt.Println(exitCode)
//   fmt.Println(err)

//   output, exitCode, err = ds.Run("alpine", []string{"cat", "x"})
//   fmt.Println(output)
//   fmt.Println(exitCode)
//   fmt.Println(err)
// }
