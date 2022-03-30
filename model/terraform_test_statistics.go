package model

import (
	"context"
	"fmt"
	"github.com/sirupsen/logrus"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"os"
	"strings"
	"sync"
	"time"
)

// TerraformTestStatistics
type TerraformTestStatistics struct {
	Id           int64     `gorm:"column:id"`            // ID
	ResourceName string    `gorm:"column:resource_name"` // 资源名
	CloudProduct string    `gorm:"column:cloud_product"` // 云产品
	Tag          string    `gorm:"column:tag"`           // 标杆客户
	CodeCoverage float64   `gorm:"column:code_coverage"` // 云产品代码码覆盖率
	Extension    string    `gorm:"column:extension"`     // 扩展，备用
	GmtCreated   time.Time `gorm:"column:gmt_created"`   // 创建日期
	GmtModified  time.Time `gorm:"column:gmt_modified"`  // 更新日期
}

func (TerraformTestStatistics) TableName() string {
	return "terraform_test_statistics"
}

// TerraformTestStatisticsDao 数据访问对象
type TerraformTestStatisticsDao struct {
}

var terraformTestStatisticsDao *TerraformTestStatisticsDao
var terraformTestStatisticsOnce sync.Once

// NewTerraformTestStatisticsDaoInstance TerraformTestStatisticsDaoInstance
func NewTerraformTestStatisticsDaoInstance() *TerraformTestStatisticsDao {
	terraformTestStatisticsOnce.Do(
		func() {
			terraformTestStatisticsDao = &TerraformTestStatisticsDao{}
		})
	return terraformTestStatisticsDao
}

// CreateResource
func (d *TerraformTestStatisticsDao) CreateResourceRecord(ctx context.Context, db *gorm.DB, item *TerraformTestStatistics) (id int64, isDuplicated bool, err error) {
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
		logrus.Error(ctx, "CreateResourceRecord failed, rowsAffected: %d, err: %v", db.RowsAffected, db.Error)
		return -1, false, fmt.Errorf("CreateResourceRecord failed")
	}
	return item.Id, false, nil
}

// GetWriteDB 可读可写DB
func GetWriteDB(ctx context.Context) *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s)/terraform_test?charset=utf8mb4&parseTime=True&loc=Local", os.Getenv("DB_Account"), os.Getenv("DB_Password"), os.Getenv("DB_EndPoint"))
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
	sqlDB, err := db.DB()
	sqlDB.SetMaxIdleConns(10)
	sqlDB.SetMaxOpenConns(100)
	sqlDB.SetConnMaxLifetime(time.Hour)
	if err != nil {
		logrus.Error(err)
		return nil
	}
	if err != nil {
		logrus.Error(err)
		return nil
	}

	return db
}
