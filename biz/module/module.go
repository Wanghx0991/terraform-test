package module

import (
	"errors"
	"fmt"
	"github.com/aliyun/terraform-test/common/util"
	"github.com/aliyun/terraform-test/consts"
	"github.com/goinggo/mapstructure"
	"github.com/imdario/mergo"
	"github.com/sirupsen/logrus"
	"os"
	"regexp"
	"strings"
	"sync"
)

var (
	wg sync.WaitGroup
	//Make channels to pass fatal errors in WaitGroup
	fatalErrors chan error
	wgDone      chan bool
)

func init() {
	fatalErrors = make(chan error)
	wgDone = make(chan bool)
}

type Module struct {
	ID              string                   `json:"id"`
	Owner           string                   `json:"owner"`
	Namespace       string                   `json:"namespace"`
	Name            string                   `json:"name"`
	Version         string                   `json:"version"`
	Provider        string                   `json:"provider"`
	ProviderLogoUrl string                   `json:"provider_logo_url"`
	Description     string                   `json:"description"`
	Source          string                   `json:"source"`
	Tag             string                   `json:"tag"`
	PublishedAt     int                      `json:"published_at"`
	Downloads       int                      `json:"downloads"`
	Verified        bool                     `json:"verified"`
	Resources       []string                 `json:"resources"`
	Examples        []map[string]interface{} `json:"examples"`
}

func ExecuteModules(resourceName []string, stage string) (map[string][]map[string]interface{}, error) {
	defer close(fatalErrors)
	if stage != "PrevStage" && stage != "NextStage" && stage != "NewVersion" {
		return nil, errors.New("please input the correct stage. the valid values: `PrevStage`, `NextStage`, `NewVersion`")
	}
	res, modules := make(map[string][]map[string]interface{}, 0), make(map[string]Module, 0)
	var err error
	// new resource && the step1 of the existed resource
	modules, err = QueryTerraformModule(resourceName)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}

	for _, mod := range modules {
		for _, raw := range mod.Examples {
			wg.Add(1)
			logrus.Infof("Executed Module name=%s, Examples name = example/complete ", raw["module_name"])
			go ExecuteSingleModule(mod, raw, stage)
		}
	}
	// Important final goroutine to wait until WaitGroup is done
	go func() {
		wg.Wait()
		close(wgDone)
	}()
	var errArray []error
	//Wait until either WaitGroup is done
	// if error occurred, the process won't stop until all goroutine is done.
	select {
	case <-wgDone:
		// carry on
		break
	case err := <-fatalErrors:
		errArray = append(errArray, err)
		logrus.Error(err)
	}
	for _, e := range errArray {
		logrus.Fatal(e.Error())
	}

	// Operate Success
	return res, nil
}

func ExecuteSingleModule(mod Module, obj map[string]interface{}, stage string) {
	defer wg.Done()
	t := strings.Split(mod.ID, "/")
	name := strings.Split(mod.Source, "/")
	sourcePrefixName := strings.Join(t[:len(t)-1], "/")
	folderName := strings.Join([]string{name[len(name)-1], mod.Name, obj["path"].(string)}, "_")
	content := dependency(mod.Name, obj["path"].(string), sourcePrefixName, stage)
	folderName = strings.ReplaceAll(folderName, "/", "_")
	err := processDir(folderName, content)
	if err != nil {
		logrus.Error(err)
		fatalErrors <- err
		return
	}
}

func QueryTerraformModule(resourceTarget []string) (map[string]Module, error) {
	objects := make(map[string]Module, 0)
	client := new(util.Client)
	first := true
	client.Query = map[string]string{
		"provider": "alicloud",
	}

	var meta, resp map[string]interface{}
	var err error
	for {
		if first {
			client.RequestPath = consts.TerrafromBaseUrl
			resp, err = client.Get()
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
			//if v, exist := resp["modules"]; exist {
			//	m, err := querySpecifiedResources(v.([]interface{}), resourceTarget)
			//	if err != nil {
			//		logrus.Error()
			//		return nil, err
			//	}
			//	err = mergo.Merge(&objects, m)
			//	if err != nil {
			//		logrus.Error()
			//		return nil, err
			//	}
			//}

			meta = resp["meta"].(map[string]interface{})
			first = false
			continue
		}

		url, exist := meta["next_url"]
		if !exist {
			break
		}
		urlRegex := regexp.MustCompile("^https:\\/\\/\\/([a-z?&/0-9=]*)")
		Matched := urlRegex.FindAllStringSubmatch(url.(string), -1)
		client.RequestPath = consts.TerraformUrl + Matched[len(Matched)-1][1]
		resp, err = client.Get()
		if err != nil {
			logrus.Error(err)
			return nil, err
		}

		if v, exist := resp["modules"]; exist {
			m, err := querySpecifiedResources(v.([]interface{}), resourceTarget)
			if err != nil {
				logrus.Error()
				return nil, err
			}
			err = mergo.Merge(&objects, m)
			if err != nil {
				logrus.Error()
				return nil, err
			}
		}

		meta = resp["meta"].(map[string]interface{})
		first = false
	}
	return objects, nil
}

func querySpecifiedResources(arr []interface{}, resourceTarget []string) (map[string]Module, error) {
	client := new(util.Client)
	//clientGithub := github.NewClient(oauth2.NewClient(context.Background(), oauth2.StaticTokenSource(
	//	&oauth2.Token{AccessToken: os.Getenv("GITHUB_TOKEN")},
	//)))
	objects := make(map[string]Module, 0)
	for _, raw := range arr {
		var mod Module
		if err := mapstructure.Decode(raw, &mod); err != nil {
			logrus.Error(err)
			return nil, err
		}
		if mod.Namespace == "terraform-alicloud-modules" {
			client.RequestPath = fmt.Sprintf("%s/%s/%s/%s", consts.TerrafromBaseUrl, mod.Namespace, mod.Name, mod.Provider)
			moduleInfo, err := client.Get()
			if err != nil {
				logrus.Error(err)
				return nil, err
			}
			flag := false
			for _, resource := range moduleInfo["root"].(map[string]interface{})["resources"].([]interface{}) {
				resourceType := resource.(map[string]interface{})["type"]
				mod.Resources = append(mod.Resources, fmt.Sprint(resourceType))
				for _, targetResource := range resourceTarget {
					if resourceType == targetResource {
						flag = true
						break
					}
				}
			}
			if !flag {
				continue
			}
			var exampleTF map[string]interface{}
			for _, example := range moduleInfo["examples"].([]interface{}) {
				if example.(map[string]interface{})["name"] == "complete" {
					exampleTF = example.(map[string]interface{})
				}
			}

			//_, examplesContent, _, err := clientGithub.Repositories.GetContents(context.Background(), mod.Namespace, fmt.Sprintf("terraform-alicloud-%s", mod.Name), "./examples/complete", nil)
			if err != nil {
				if util.IsExpectedErrors(err, []string{"Not Found"}) {
					continue
				}
				return nil, errors.New(fmt.Sprintf("Query terraform-alicloud-%s error", mod.Name))
			}

			mod.Examples = append(mod.Examples, map[string]interface{}{
				"module_name": fmt.Sprintf("terraform-alicloud-%s", mod.Name),
				"name_space":  mod.Namespace,
				"path":        "examples/complete",
				"exampleTF":   exampleTF,
			})

			// 涉及到相关资源, 且examples不为空
			if flag && exampleTF != nil {
				//if flag && len(examplesContent) != 0 && exampleTF != nil {
				logrus.Infof("==== The Relative Module: terraform-alicloud-%s ====", mod.Name)
				objects[mod.Name] = mod
			}
			flag = false
			if len(mod.Resources) == 0 {
				logrus.Warningf("The resource is empty under the namespace: %s, name = %s, resource number = %d", mod.Namespace, mod.Name, len(mod.Resources))
			}
		}
	}
	return objects, nil
}

func dependency(moduleName, examplePath, sourcePrefixName, stage string) string {
	config := fmt.Sprintf(`
module "%s"  {
	source  = "%s//%s"
}
	`, moduleName, sourcePrefixName, examplePath)
	if stage == "NextStage" || stage == "NewVersion" {
		config += fmt.Sprint(`
terraform {
  required_providers {
    alicloud = {
      source  = "terraform.local/local/alicloud"
      version = "1.0.0"
    }
  }
}
`)
	}

	return config
}

func processDir(dirname, content string) error {
	err := os.MkdirAll("./tmp/"+dirname, 0777)
	if err != nil {
		logrus.Error(err)
		return err
	}
	f, err := os.Create("./tmp/" + dirname + "/main.tf")
	defer f.Close()
	if err != nil {
		logrus.Error(err)
		return err
	}

	_, err = f.WriteString(content)
	if err != nil {
		logrus.Error(err)
		return err
	}
	return nil
}
