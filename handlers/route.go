// Copyright (c) Autovia GmbH
// SPDX-License-Identifier: Apache-2.0

package handlers

import (
	"log"
	"net/http"

	S "github.com/autovia/dyn/structs"
)

func Post(a *S.App, w http.ResponseWriter, req *http.Request) error {
	log.Printf(">>> POST %v\n", req)

	switch req.Header.Get("X-Amz-Target") {
	case "AmazonEC2ContainerServiceV20141113.RegisterTaskDefinition":
		return RegisterTaskDefinition(a, w, req)
	case "AmazonEC2ContainerServiceV20141113.RunTask":
		return RunTask(a, w, req)
	case "AmazonEC2ContainerServiceV20141113.ListTaskDefinitions":
		return ListTaskDefinitions(a, w, req)
	}

	return nil
}
