package main

import (
	"fmt"
	"github.com/aliyun/terraform-test/biz/module"
	"github.com/sirupsen/logrus"
	"testing"
)

func TestQueryTerraform(t *testing.T) {
	mod, err := module.ExecuteModules([]string{"alicloud_slb_listener"}, "")
	fmt.Println(mod)
	if err != nil {
		logrus.Error(err)
		t.Error()
	}
}
