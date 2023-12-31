// Copyright 2022 The kubegems.io Authors
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

package myinfohandler

import (
	"github.com/gin-gonic/gin"
	"kubegems.io/kubegems/pkg/service/handlers/base"
)

type MyHandler struct {
	base.BaseHandler
}

func (h *MyHandler) RegistRouter(rg *gin.RouterGroup) {
	rg.GET("/my/info", h.Myinfo)
	rg.GET("/my/auth", h.MyAuthority)
	rg.GET("/my/tenants", h.MyTenants)
	rg.POST("/my/reset_password", h.ResetPassword)
}
