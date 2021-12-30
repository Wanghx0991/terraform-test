package code_coverage

import (
	"bytes"
	"context"
	"fmt"
	"github.com/PuerkitoBio/goquery"
	"github.com/aliyun/aliyun-oss-go-sdk/oss"
	"github.com/aliyun/terraform-test/common/util"
	"github.com/aliyun/terraform-test/consts"
	"github.com/aliyun/terraform-test/model"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var flag = map[string][]string{
	"cbn":                {"cen", "cbn"},
	"ecs":                {"router", "ecs", "instance", "security_group", "image"},
	"cr":                 {"cr"},
	"ram":                {"ram"},
	"alidns":             {"alidns"},
	"sddp":               {"sddp"},
	"dm":                 {"direct_mail", "dm"},
	"cas":                {"ssl_certificates", "cas"},
	"arms":               {"arms"},
	"cs":                 {"cs"},
	"elasticsearch":      {"elasticsearch"},
	"rds":                {"db", "rds"},
	"amqp":               {"amqp"},
	"sas":                {"sas", "security_center_group"},
	"vpc":                {"vpc", "eip", "vswitch", "route"},
	"r-kvstore":          {"kvstore", "r_kvstore"},
	"actiontrail":        {"actiontrail"},
	"cms":                {"cms"},
	"yundun-bastionhost": {"yundun_bastionhost", "bastionhost"},
	"slb":                {"slb"},
	"waf-openapi":        {"waf"},
	"cdn":                {"cdn", "scdn"},
	"cloudfw":            {"cloudfw", "cloud_firewall"},
}

type CodeCoverageHandler struct {
	SourceFilePath string
	TargetPath     string
}

type ResourceRes struct {
	index            string
	fileName         string
	cloudProduct     string
	CodeCoverageRate float64
	Tag              []string
}

func (s *CodeCoverageHandler) ConvertFile(filename string) (string, error) {
	_, _, err := util.DoCmd("cd ../../tmp/ && rm -rf ./terraform-provider-alicloud && git clone https://github.com/aliyun/terraform-provider-alicloud && cd ./terraform-provider-alicloud")
	if err != nil {
		logrus.Error(err)
	}
	client, err := oss.New(consts.OssEndpointBeijing, os.Getenv("ALICLOUD_ACCESS_KEY"), os.Getenv("ALICLOUD_SECRET_KEY"))
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	bucket, err := client.Bucket("terraform-ci")
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	err = bucket.GetObjectToFile(fmt.Sprintf("coverage/eu-central-1/%s", filename), "../../tmp/terraform-provider-alicloud/All.out")
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	command := fmt.Sprintf("cd ../../tmp/terraform-provider-alicloud && go tool cover -html %s -o %s", "All.out", "All.html")
	res, _, err := util.DoCmd(command)
	if err != nil {
		logrus.Error(err)
		return "", err
	}
	return res, nil
}

func (s *CodeCoverageHandler) ParseHtmlFile() ([]ResourceRes, error) {
	file, err := ioutil.ReadFile(s.TargetPath) //这个就是读取你的html
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	doc, err := goquery.NewDocumentFromReader(bytes.NewReader(file))
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	splitRes := make([]string, 0)
	doc.Find("#files").Each(func(i int, selection *goquery.Selection) {
		text := selection.Text()
		splitRes = strings.Fields(text)
	})
	index := 0
	record := make([]ResourceRes, 0)
	Regexp := regexp.MustCompile(`github.com/aliyun/terraform-provider-alicloud/alicloud/([a-zA-Z0-9_]*).go$`)
	for i := 0; i < len(splitRes); i++ {
		if i%2 != 0 {
			continue
		}
		obj := new(ResourceRes)
		params := Regexp.FindStringSubmatch(splitRes[i])[1]
		obj.cloudProduct = processCloudProduct(params)
		obj.fileName = params
		obj.index = fmt.Sprintf("file%d", index)
		val := []byte(splitRes[i+1])
		obj.CodeCoverageRate, err = strconv.ParseFloat(string(val[1:len(val)-2]), 2)

		// 打tag
		update := false
		for _, codeArray := range flag {
			for _, code := range codeArray {
				if strings.Contains(obj.fileName, code) {
					update = true
					break
				}
			}
		}
		if update {
			obj.Tag = append(obj.Tag, "Align")
		}

		recd := &model.TerraformTestStatistics{
			ResourceName: obj.fileName,
			CodeCoverage: obj.CodeCoverageRate * 10,
			CloudProduct: obj.cloudProduct,
			Tag:          strings.Join(obj.Tag, ","),
			GmtCreated:   time.Now(),
			GmtModified:  time.Now(),
		}
		fmt.Printf("FileName = %s , CloudProduct = %s\n\n", recd.ResourceName, recd.CloudProduct)
		model.NewTerraformTestStatisticsDaoInstance().CreateResourceRecord(context.Background(), nil, recd)
		index++
		record = append(record, *obj)
	}
	return record, nil
}

func processCloudProduct(origin string) (cloudProduct string) {
	if strings.Contains(origin, "data_source_alicloud") && len(strings.Split(origin, "_")) >= 3 {
		split := strings.Split(origin, "_")
		if split[3] == "cloud" {
			if split[4] == "firewall" || split[4] == "sso" || split[4] == "network" || split[4] == "connect" {
				cloudProduct = split[3] + split[4]
				return cloudProduct
			}
			if split[4] == "storage" {
				cloudProduct = "cloud_storage_gateway"
				return cloudProduct
			}
		}
		if split[3] == "cr" {
			if split[4] == "ee" {
				cloudProduct = "cr_ee"
				return cloudProduct
			}
			cloudProduct = "cr"
			return cloudProduct
		}
		if split[3] == "data" {
			if split[4] == "works" {
				cloudProduct = "dataworks"
				return cloudProduct
			}
			cloudProduct = "cr"
			return cloudProduct
		}

		if split[3] == "direct" {
			if split[4] == "mail" {
				cloudProduct = "direct_mail"
				return cloudProduct
			}
			cloudProduct = "cr"
			return cloudProduct
		}
		if split[3] == "event" {
			if split[4] == "bridge" {
				cloudProduct = "event_bridge"
				return cloudProduct
			}
			cloudProduct = "cr"
			return cloudProduct
		}

		if split[3] == "express" {
			if split[4] == "connect" {
				cloudProduct = "express_connect"
				return cloudProduct
			}
			cloudProduct = "cr"
			return cloudProduct
		}

		if v, exist := consts.NameTransfer[split[3]]; exist {
			cloudProduct = v
			return cloudProduct
		}
		cloudProduct = split[3]
		return cloudProduct
	} else if strings.Contains(origin, "resource_alicloud") && len(strings.Split(origin, "_")) >= 2 {
		split := strings.Split(origin, "_")
		if split[2] == "cloud" {
			if split[3] == "firewall" || split[3] == "sso" || split[3] == "network" || split[3] == "connect" {
				cloudProduct = split[2] + split[3]
				return cloudProduct
			}
			if split[3] == "storage" {
				cloudProduct = "cloud_storage_gateway"
				return cloudProduct
			}
		}
		if split[2] == "cr" {
			if len(split) >= 4 && split[3] == "ee" {
				cloudProduct = "cr_ee"
				return cloudProduct
			}
			cloudProduct = "cr"
			return cloudProduct
		}
		if split[2] == "express" {
			if len(split) >= 4 && split[3] == "connect" {
				cloudProduct = "express_connect"
				return cloudProduct
			}
			cloudProduct = "cr"
			return cloudProduct
		}
		if split[2] == "event" {
			if len(split) >= 4 && split[3] == "bridge" {
				cloudProduct = "event_bridge"
				return cloudProduct
			}
			cloudProduct = "cr"
			return cloudProduct
		}
		if v, exist := consts.NameTransfer[split[2]]; exist {
			cloudProduct = v
			return cloudProduct
		}
		cloudProduct = split[2]
		return cloudProduct
	} else if strings.Contains(origin, "service_alicloud") && len(strings.Split(origin, "_")) >= 2 {
		split := strings.Split(origin, "_")
		if split[2] == "r" {
			cloudProduct = "kvstore"
			return cloudProduct
		}
		cloudProduct = strings.Split(origin, "_")[2]
		return cloudProduct
	}
	return cloudProduct
}
