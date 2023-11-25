// Copyright (c) Autovia GmbH
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/containerd/containerd"
	"github.com/containerd/containerd/cio"
	"github.com/containerd/containerd/namespaces"
	"github.com/containerd/containerd/oci"

	S "github.com/autovia/dyn/structs"
)

func RegisterTaskDefinition(app *S.App, w http.ResponseWriter, req *http.Request) error {
	log.Printf("#RegisterTaskDefinition %v\n", req)

	body, _ := io.ReadAll(req.Body)

	var taskIn *ecs.RegisterTaskDefinitionInput
	log.Print(string(body))

	err := json.Unmarshal(body, &taskIn)

	if err != nil {
		fmt.Println(err)
	}

	path := filepath.Join(*app.Mount, *app.Metadata, *taskIn.Family)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err := os.Mkdir(path, os.ModePerm); err != nil {
			return app.RespondError(w, 500, "InternalError", err, "Mkdir")
		}
	}

	contents, err := os.ReadDir(path)
	if err != nil {
		return app.RespondError(w, http.StatusInternalServerError, "InternalError", err, "ReadDir")
	}

	var revision int
	if len(contents) > 0 {
		for _, file := range contents {
			fileInfo, _ := file.Info()
			s, err := strconv.Atoi(fileInfo.Name())
			if err != nil {
				fmt.Println("Can't convert filename to an int!")
			} else {
				if s > revision {
					revision = s
				}
			}
		}
	}

	revisionFile, err := os.Create(filepath.Join(path, fmt.Sprintf("%v", revision+1)))
	if err != nil {
		return app.RespondError(w, http.StatusInternalServerError, "InternalError", err, "")
	}
	defer revisionFile.Close()
	_, err = revisionFile.Write(body)
	if err != nil {
		return app.RespondError(w, http.StatusInternalServerError, "InternalError", err, "")
	}

	ctx := namespaces.WithNamespace(context.Background(), "dyn")
	for _, c := range taskIn.ContainerDefinitions {
		go pullImage(ctx, app.C, *c.Image)
	}

	images, err := app.C.ListImages(ctx, "")
	if err != nil {
		return err
	}

	for _, v := range images {
		log.Print(">>> ", v)
	}

	taskOut := ecs.RegisterTaskDefinitionOutput{
		TaskDefinition: &ecs.TaskDefinition{
			ContainerDefinitions: taskIn.ContainerDefinitions,
			Revision:             i(int64(revision + 1)),
			Status:               s("ACTIVE"),
			TaskDefinitionArn:    s(fmt.Sprintf("%s:%v", *taskIn.Family, revision+1)),
		},
	}

	return app.RespondECS(w, http.StatusOK, taskOut)
}

func pullImage(ctx context.Context, client *containerd.Client, name string) {
	img := "docker.io/library/" + name
	if !strings.Contains(name, ":") {
		img += ":latest"
	}

	image, err := client.Pull(ctx, img, containerd.WithPullUnpack)

	if err != nil {
		log.Print(err)
	}
	log.Print(image)
}

func s(s string) *string {
	return &s
}

func i(i int64) *int64 {
	return &i
}

func ListTaskDefinitions(app *S.App, w http.ResponseWriter, req *http.Request) error {
	log.Printf("#ListTaskDefinitions %v\n", req)

	body, _ := io.ReadAll(req.Body)
	var listIn *ecs.ListTaskDefinitionsInput
	log.Print(string(body))
	err := json.Unmarshal(body, &listIn)
	if err != nil {
		fmt.Println(err)
	}

	log.Print("....", string(body))

	path := filepath.Join(*app.Mount, *app.Metadata)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return app.RespondError(w, 500, "InternalError", err, "IsNotExist")
	}

	listOut := ecs.ListTaskDefinitionsOutput{}
	_ = filepath.Walk(path, func(p string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Println(err)
			return err
		}
		if !info.IsDir() {
			dir, file := filepath.Split(p)
			def, _ := strings.CutPrefix(dir, path)
			log.Println(file)
			n := strings.ReplaceAll(def, "/", "") + ":" + file
			listOut.TaskDefinitionArns = append(listOut.TaskDefinitionArns, &n)

		}
		return nil
	})
	if err != nil {
		fmt.Println(err)
	}

	return app.RespondECS(w, http.StatusOK, listOut)
}

func RunTask(app *S.App, w http.ResponseWriter, req *http.Request) error {
	log.Printf("#RunTask %v\n", req)

	body, _ := io.ReadAll(req.Body)

	var runIn *ecs.RunTaskInput
	log.Print(string(body))
	err := json.Unmarshal(body, &runIn)
	if err != nil {
		fmt.Println(err)
	}

	log.Print(*runIn.TaskDefinition)
	def := strings.Split(*runIn.TaskDefinition, ":")
	if len(def) != 2 {
		return err
	}

	path := filepath.Join(*app.Mount, *app.Metadata, def[0], def[1])

	f, err := os.Open(path)
	if err != nil {
		fmt.Println(err)
	}
	content, err := io.ReadAll(f)
	if err != nil {
		fmt.Println(err)
	}
	var taskDef *ecs.TaskDefinition
	err = json.Unmarshal(content, &taskDef)
	if err != nil {
		fmt.Println(err)
	}

	ctx := namespaces.WithNamespace(context.Background(), "dyn")
	for _, c := range taskDef.ContainerDefinitions {
		startContainer(ctx, app.C, c)
	}

	return nil
}

func startContainer(ctx context.Context, client *containerd.Client, container *ecs.ContainerDefinition) {
	if len(container.Command) > 0 {
		var cmd []string

		for i := 0; i < len(container.Command); i++ {
			cmd = append(cmd, fmt.Sprintf("%v", *container.Command[i]))
		}

		fmt.Println(cmd)

		c, _ := client.LoadContainer(ctx, *container.Name)
		if c == nil {
			img := "docker.io/library/" + *container.Image
			if !strings.Contains(*container.Image, ":") {
				img += ":latest"
			}

			image, err := client.GetImage(ctx, img)
			if err != nil {
				fmt.Println(err)
			}

			c, err = client.NewContainer(ctx, *container.Name,
				containerd.WithNewSnapshot(*container.Name, image),
				containerd.WithNewSpec(oci.WithImageConfig(image),
					oci.WithProcessArgs(cmd...)))
			if err != nil {
				fmt.Println(err)
			}
		}

		task, err := c.NewTask(ctx, cio.NewCreator(cio.WithStdio))
		if err != nil {
			fmt.Println(err)
		}
		fmt.Println(task)
		defer task.Delete(ctx)

		exitStatusC, err := task.Wait(ctx)
		if err != nil {
			fmt.Println(err)
		}

		if err := task.Start(ctx); err != nil {
			fmt.Println(err)
		}

		status := <-exitStatusC
		code, _, err := status.Result()
		fmt.Println(code)

		if err != nil {
			fmt.Println(err)
		}
		fmt.Printf("status: %d\n", code)
	}

}
