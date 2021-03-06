package code_coverage

import (
	"fmt"
	"github.com/aliyun/terraform-test/model"
	"github.com/sirupsen/logrus"
	"golang.org/x/net/context"
	"os"
	"testing"
	"time"
)

func TestFileConvert(t *testing.T) {
	handler := &CodeCoverageHandler{}
	res, err := handler.ConvertFile("All-2021-11-09.out")
	if err != nil {
		t.Error(err)
	}
	logrus.Info(res)
}

func TestParseHtml(t *testing.T) {
	handler := &CodeCoverageHandler{
		TargetPath: "../../tmp/terraform-provider-alicloud/All.html",
	}
	_, err := handler.ParseHtmlFile()
	if err != nil {
		t.Error(err)
	}
}

func TestEnv(t *testing.T) {
	st, _ := os.LookupEnv("DB_Account")
	fmt.Println(st)
}

func TestInsertRecord(t *testing.T) {
	Tim, _ := time.Parse("2006-01-02 15:04:05", "2021-11-14 00:00:00")
	obj := &model.TerraformTestSummary{
		Id:           0,
		SuccessRate:  77.1 * 10,
		CodeCoverage: 52.3 * 10,
		GmtCreated:   Tim,
		Extension:    "",
	}
	model.NewTerraformTestSummaryDaoInstance().CreateResourceRecord(context.Background(), nil, obj)
}
