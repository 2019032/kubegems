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

package models

import (
	"time"
)

type Announcement struct {
	ID      uint       `gorm:"primarykey" json:"id"`
	Type    string     `gorm:"type:varchar(50);" json:"type"`
	Message string     `json:"message"`
	StartAt *time.Time `json:"startAt"` // 开始时间，默认现在
	EndAt   *time.Time `json:"endAt"`   // 结束时间，默认一天后

	CreatedAt *time.Time `json:"createdAt"`
	UpdatedAt *time.Time `json:"updatedAt"`
}
