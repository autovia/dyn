// Copyright (c) Autovia GmbH
// SPDX-License-Identifier: Apache-2.0

package structs

import (
	"net/http"

	"github.com/containerd/containerd"
)

type App struct {
	Addr      *string
	Router    *http.ServeMux
	AccessKey *string
	SecretKey *string
	Mount     *string
	Metadata  *string
	C         *containerd.Client
}
