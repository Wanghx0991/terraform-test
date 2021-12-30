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

// TerraformTestStatistics
type TerraformTestSummary struct {
	Id           int64     `gorm:"column:id"`            // ID
	SuccessRate  float64   `gorm:"column:success_rate"`  // 测试成功率
	CodeCoverage float64   `gorm:"column:code_coverage"` // 云产品代码码覆盖率
	Extension    string    `gorm:"column:extension"`     // 扩展，备用
	GmtCreated   time.Time `gorm:"column:gmt_created"`   // 创建日期
}

func (TerraformTestSummary) TableName() string {
	return "terraform_test_summary"
}

// TerraformTestSummaryDao 数据访问对象
type TerraformTestSummaryDao struct {
}

var terraformTestSummaryDaoDao *TerraformTestSummaryDao
var terraformTestSummaryDaoDaoOnce sync.Once

// NewTerraformTestStatisticsDaoInstance TerraformTestStatisticsDaoInstance
func NewTerraformTestSummaryDaoInstance() *TerraformTestSummaryDao {
	terraformTestSummaryDaoDaoOnce.Do(
		func() {
			terraformTestSummaryDaoDao = &TerraformTestSummaryDao{}
		})
	return terraformTestSummaryDaoDao
}

// CreateResource
func (d *TerraformTestSummaryDao) CreateResourceRecord(ctx context.Context, db *gorm.DB, item *TerraformTestSummary) (id int64, isDuplicated bool, err error) {
	if db == nil {
		db = GetWriteDB(ctx)
	}
	db = db.Create(item)
	if db.Error != nil {
		if strings.Contains(db.Error.Error(), "Duplicate") {
			// 重复插入
			logrus.Warn(ctx, "CreateResourceRecord duplicated, rowsAffected: %d, err: %v", db.RowsAffected, db.Error)
			return -1, true, db.Error
		}
		logrus.Error(ctx, "CreateResourceRecord err: %v", db.Error)
		return -1, false, db.Error
	}
	if db.RowsAffected != 1 { // 数据库无错误但是没有插入成功
		logrus.Error(ctx, "TerraformTestSummary failed, rowsAffected: %d, err: %v", db.RowsAffected, db.Error)
		return -1, false, fmt.Errorf("TerraformTestSummary failed")
	}
	return item.Id, false, nil
}
