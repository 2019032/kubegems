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
	"fmt"
	"os"
	"path"
	"time"

	"gorm.io/gorm"
	"kubegems.io/kubegems/pkg/utils"
	"kubegems.io/kubegems/pkg/utils/loki"
)

// LogQueryHistory 日志查询历史
type LogQueryHistory struct {
	ID uint `gorm:"primarykey"`
	// 关联的集群
	Cluster *Cluster `gorm:"constraint:OnUpdate:RESTRICT,OnDelete:CASCADE;"`
	// 所属集群ID
	ClusterID uint
	// 标签
	LabelJSON string `gorm:"type:varchar(1024)"`
	// 正则标签
	FilterJSON string `gorm:"type:varchar(1024)"`
	// logql
	LogQL string `gorm:"type:varchar(1024)"`
	// 时间区间
	TimeRange string `gorm:"type:varchar(256);default:''"`
	// 创建时间
	CreateAt time.Time `sql:"DEFAULT:'current_timestamp'"`
	// 创建者
	Creator   *User `gorm:"constraint:OnUpdate:RESTRICT,OnDelete:CASCADE;"`
	CreatorID uint
}

type LogQueryHistoryWithCount struct {
	ID         uint
	Ids        string
	TimeRanges string
	Cluster    *Cluster
	ClusterID  uint
	LabelJSON  string
	FilterJSON string
	LogQL      string
	CreateAt   time.Time
	Creator    *User
	CreatorID  uint
	Total      string
}

// LogQuerySnapshot 日志查询快照
type LogQuerySnapshot struct {
	ID uint `gorm:"primarykey"`
	// 关联的集群
	Cluster *Cluster `gorm:"constraint:OnUpdate:RESTRICT,OnDelete:CASCADE;"`
	// 所属集群ID
	ClusterID uint
	// 名称
	SnapshotName string `gorm:"type:varchar(128)"`
	SourceFile   string `gorm:"type:varchar(128)"`
	// 行数
	SnapshotCount int
	// 下载地址
	DownloadURL string `gorm:"type:varchar(512)"`
	StartTime   time.Time
	EndTime     time.Time
	// 创建时间
	CreateAt time.Time `sql:"DEFAULT:'current_timestamp'"`
	// 创建者
	Creator   *User `gorm:"constraint:OnUpdate:RESTRICT,OnDelete:CASCADE;"`
	CreatorID uint
}

func (snapshot *LogQuerySnapshot) BeforeCreate(tx *gorm.DB) error {
	var (
		lineCount int64
		err       error
	)
	lokiExportDir := "data/lokiExport"

	lokiSnapshotDir := path.Join(lokiExportDir, "snapshot", time.Now().UTC().Format("20060102"))
	err = utils.EnsurePathExists(lokiSnapshotDir)
	if err != nil {
		return err
	}
	sourceFile := path.Join(lokiExportDir, snapshot.SourceFile)
	if loki.FileExists(sourceFile) {
		targetFile := path.Join(lokiSnapshotDir, snapshot.SnapshotName)
		if !loki.FileExists(targetFile) {
			lineCount, err = utils.CopyFileByLine(targetFile, sourceFile)
			if err != nil {
				return err
			}
		} else {
			return fmt.Errorf("名字为 %s 的日志快照 已经存在，请换个名字保存", snapshot.SnapshotName)
		}
		snapshot.SnapshotCount = int(lineCount)
		snapshot.DownloadURL = path.Join("/", lokiSnapshotDir, snapshot.SnapshotName)
		return nil
	}

	return nil
}

func (snapshot *LogQuerySnapshot) BeforeDelete(tx *gorm.DB) error {
	if snapshot.DownloadURL != "" && loki.FileExists(snapshot.DownloadURL) {
		err := os.Remove(snapshot.DownloadURL)
		if err != nil {
			return err
		}
	}
	return nil
}
