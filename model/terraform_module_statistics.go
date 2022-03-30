package model

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/gorm"
	"strings"
	"sync"
	"time"
)

// TerraformModuleStatistics TerraformTestStatistics
type TerraformModuleStatistics struct {
	Id         int64     `gorm:"column:id"`          // ID
	Namespace  string    `gorm:"column:namespace"`   // 命名空间
	ModuleName string    `gorm:"column:module_name"` // module
	Version    string    `gorm:"column:version"`     // 标杆客户
	Verified   string    `gorm:"column:verified"`    //
	Tag        string    `gorm:"column:tag"`         //
	Source     string    `gorm:"column:source"`      // 仓库地址
	Resources  string    `gorm:"column:resources"`   // 涉及资源
	Examples   string    `gorm:"column:examples"`    // Examples
	GmtCreated time.Time `gorm:"column:gmt_created"` // 创建日期
}

func (TerraformModuleStatistics) TableName() string {
	return "terraform_module_statistics"
}

// TerraformModuleStatisticsDao 数据访问对象
type TerraformModuleStatisticsDao struct {
}

var terraformModuleStatisticsDao *TerraformModuleStatisticsDao
var terraformModuleStatisticsOnce sync.Once

// NewTerraformModuleStatisticsDaoInstance NewTerraformTestStatisticsDaoInstance TerraformTestStatisticsDaoInstance
func NewTerraformModuleStatisticsDaoInstance() *TerraformModuleStatisticsDao {
	terraformTestStatisticsOnce.Do(
		func() {
			terraformModuleStatisticsDao = &TerraformModuleStatisticsDao{}
		})
	return terraformModuleStatisticsDao
}

// CreateModuleRecord Create
func (d *TerraformModuleStatisticsDao) CreateModuleRecord(ctx context.Context, db *gorm.DB, item *TerraformModuleStatistics) (id int64, isDuplicated bool, err error) {
	if db == nil {
		db = GetWriteDB(ctx)
	}
	db = db.Create(item)
	if db.Error != nil {
		if strings.Contains(db.Error.Error(), "Duplicate") {
			// 重复插入
			logrus.Warn(ctx, "CreateModuleCreate duplicated, rowsAffected: %d, err: %v", db.RowsAffected, db.Error)
			return -1, true, db.Error
		}
		logrus.Error(ctx, "CreateModuleCreate err: %v", db.Error)
		return -1, false, db.Error
	}
	if db.RowsAffected != 1 { // 数据库无错误但是没有插入成功
		logrus.Error(ctx, "CreateResourceRecord failed, rowsAffected: %d, err: %v", db.RowsAffected, db.Error)
		return -1, false, fmt.Errorf("CreateResourceRecord failed")
	}
	return item.Id, false, nil
}
