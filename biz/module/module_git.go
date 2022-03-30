package module

import (
	"context"

	"encoding/json"
	"fmt"
	"github.com/aliyun/terraform-test/common/util"
	"github.com/aliyun/terraform-test/consts"
	"github.com/aliyun/terraform-test/model"
	"github.com/sirupsen/logrus"
	"math"
	"strings"
	"time"
)

type TFModule struct {
	Name string
}

func queryModule() (interface{}, error) {
	iter := int(math.Ceil(consts.ModulesNume / consts.PerPage))
	if consts.ModulesNume%consts.PerPage != 0 {
		iter++
	}
	count := 0
	for i := 1; i <= iter; i++ {
		client := new(util.Client)
		RequestPath := fmt.Sprintf("%s?page=%d&per_page=%d", consts.TerraformModuleRepoUrl, i, consts.PerPage)
		client.RequestPath = RequestPath
		resp, err := client.GetCommon()
		if err != nil {
			logrus.Error(err)
			return nil, err
		}
		contentArray := make([]interface{}, 0)
		json.Unmarshal(resp.([]byte), &contentArray)

		for index, raw := range contentArray {
			obj := raw.(map[string]interface{})
			fmt.Println(index+(i-1)*10, obj["name"])
			record := model.TerraformModuleStatistics{
				Id:         0,
				Namespace:  "terraform-alicloud-modules",
				ModuleName: obj["name"].(string),
				Verified:   "NotYet",
				Source:     obj["html_url"].(string),
				GmtCreated: time.Now(),
			}
			if strings.Contains(record.ModuleName, "terraform-alicloud-") {
				res := strings.Split(record.ModuleName, "terraform-alicloud-")
				record.ModuleName = res[len(res)-1]
			}
			client.RequestPath = fmt.Sprintf("%s/%s/%s/%s", consts.TerrafromBaseUrl, "terraform-alicloud-modules", record.ModuleName, "alicloud")
			moduleInfo, err := client.Get()
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
			if _, exist := moduleInfo["errors"]; exist {
				count++
				record.ModuleName = "terraform-alicloud-" + record.ModuleName
				logrus.Infof("Modules Number: %d, Current Module: %s\n", count, record.ModuleName)
				model.NewTerraformModuleStatisticsDaoInstance().CreateModuleRecord(context.Background(), nil, &record)
				continue
			}
			record.Source = moduleInfo["source"].(string)
			record.Tag = moduleInfo["tag"].(string)
			record.ModuleName = "terraform-alicloud-" + moduleInfo["name"].(string)
			resources := make([]string, 0)
			for _, resource := range moduleInfo["root"].(map[string]interface{})["resources"].([]interface{}) {
				resourceType := resource.(map[string]interface{})["type"]
				resources = append(resources, fmt.Sprint(resourceType))
			}
			examples := make([]string, 0)
			for _, example := range moduleInfo["examples"].([]interface{}) {
				examples = append(examples, example.(map[string]interface{})["name"].(string))
			}
			v, err := json.Marshal(resources)
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
			record.Resources = string(v)
			v, err = json.Marshal(examples)
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
			record.Examples = string(v)
			count++
			model.NewTerraformModuleStatisticsDaoInstance().CreateModuleRecord(context.Background(), nil, &record)
			logrus.Infof("Modules Number: %d, Current Module: %s\n", count, record.ModuleName)

		}
	}
	logrus.Infof("Modules Number: %d", count)
	return nil, nil
}
