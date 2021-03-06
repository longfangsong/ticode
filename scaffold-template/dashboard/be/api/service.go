// Copyright 2020 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package __APP_NAME__

import (
	"net/http"
	"sort"
	"strings"

	"github.com/joomcode/errorx"

	"github.com/gin-gonic/gin"

	"go.uber.org/fx"

	"github.com/pingcap-incubator/tidb-dashboard/pkg/apiserver/user"
	"github.com/pingcap-incubator/tidb-dashboard/pkg/apiserver/utils"
	"github.com/pingcap-incubator/tidb-dashboard/pkg/tidb"
)

var (
	ErrNS     = errorx.NewNamespace("error.api.__APP_NAME__")
	ErrNoData = ErrNS.NewType("export_no_data")
)

type ServiceParams struct {
	fx.In
	TiDBClient *tidb.Client
}

type Service struct {
	params ServiceParams
}

func NewService(p ServiceParams) *Service {
	return &Service{params: p}
}

func RegisterRouter(r *gin.RouterGroup, auth *user.AuthService, s *Service) {
	endpoint := r.Group("/__APP_NAME__")
	{
		endpoint.Use(auth.MWAuthRequired())
		endpoint.Use(utils.MWConnectTiDB(s.params.TiDBClient))
		{
			endpoint.GET("/hello", s.helloHandler)
		}
	}
}

type HelloResponse struct {
	Echo      string   `json:"echo"`
	Databases []string `json:"databases"`
}

// @Summary List all databases
// @Router /__APP_NAME__/hello [get]
// @Success 200 {object} HelloResponse
// @Param name query string true "name"
// @Security JwtAuth
// @Failure 401 {object} utils.APIError "Unauthorized failure"
func (s *Service) helloHandler(c *gin.Context) {
	name := c.Query("name")
	echo := "Hello " + name

	db := utils.GetTiDBConnection(c)
	var databases []string
	err := db.Raw("show databases").Pluck("Databases", &databases).Error
	if err != nil {
		_ = c.Error(err)
		return
	}
	for i, v := range databases {
		databases[i] = strings.ToLower(v)
	}
	sort.Strings(databases)

	c.JSON(http.StatusOK, HelloResponse{
		Echo:      echo,
		Databases: databases,
	})
}
