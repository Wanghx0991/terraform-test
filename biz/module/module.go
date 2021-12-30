package module

import (
	"fmt"
	"github.com/aliyun/terraform-test/common/util"
	"github.com/aliyun/terraform-test/consts"
	"github.com/goinggo/mapstructure"
	"github.com/imdario/mergo"
	"github.com/sirupsen/logrus"
	"os"
	"strings"
	"sync"
)

var wg sync.WaitGroup

//Make channels to pass fatal errors in WaitGroup
var fatalErrors chan error
var wgDone chan bool

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

type Example struct {
	Path      string        `json:"path"`
	Empty     bool          `json:"empty"`
	Inputs    []interface{} `json:"inputs"`
	Name      string        `json:"name"`
	Outputs   []interface{} `json:"outputs"`
	Resources []interface{} `json:"resources"`
}

func ExecuteModules(resourceName []string, para string) (map[string]Example, error) {
	modules, err := QueryTerraformModule(resourceName)
	defer close(fatalErrors)
	if err != nil {
		logrus.Error(err)
		return nil, err
	}
	res := make(map[string]Example, 0)
	for _, mod := range modules {
		for _, raw := range mod.Examples {
			wg.Add(1)
			var obj Example
			if err := mapstructure.Decode(raw, &obj); err != nil {
				logrus.Error(err)
				return nil, err
			}
			logrus.Infof("Executed Module name=%s, Examples name = %s ", mod.Name, obj.Name)
			res[obj.Name] = obj
			go ExecuteSingleModule(mod, obj)
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

func ExecuteSingleModule(mod Module, obj Example) error {
	defer wg.Done()
	t := strings.Split(mod.ID, "/")
	name := strings.Split(mod.Source, "/")
	moduleName := mod.Name
	sourcePrefixName := strings.Join(t[:len(t)-1], "/")
	folderName := strings.Join([]string{name[len(name)-1], moduleName, obj.Path}, "_")
	content := dependency(moduleName, obj.Path, sourcePrefixName)
	folderName = strings.ReplaceAll(folderName, "/", "_")
	err := processDir(folderName, content)
	if err != nil {
		logrus.Error(err)
		return err
	}
	var out, stdErr string
	out, stdErr, err = util.DoCmd(fmt.Sprintf("./scripts/module.sh %s", folderName))

	logrus.Infof("The Source = %s, THe Modult Name = %s\n StdOut:\n %s, Error:\n%s", mod.Source, mod.Name, out, stdErr)
	if err != nil {
		fatalErrors <- err
		return err
	}
	if strings.Contains(stdErr, "Error:") && !strings.Contains(stdErr, "Apply complete") {
		err = fmt.Errorf(stdErr)
		fatalErrors <- err
		return err
	}
	return nil
}

func QueryTerraformModule(resourceTarget []string) (map[string]Module, error) {
	objects := make(map[string]Module, 0)
	client := new(util.Client)
	first := true
	client.Query = map[string]string{
		"provider": "alicloud",
		"limit":    consts.MaxPageSize,
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
			continue
		}

		url, exist := meta["next_url"]
		if !exist {
			break
		}
		client.RequestPath = consts.TerraformUrl + url.(string)
		moduleResp, err := client.Get()
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

		meta = moduleResp["meta"].(map[string]interface{})
		first = false
	}
	return objects, nil
}

func querySpecifiedResources(arr []interface{}, resourceTarget []string) (map[string]Module, error) {
	client := new(util.Client)
	objects := make(map[string]Module, 0)
	for _, raw := range arr {
		var mod Module
		if err := mapstructure.Decode(raw, &mod); err != nil {
			logrus.Error(err)
			return nil, err
		}
		if mod.Namespace == "terraform-alicloud-modules" || mod.Namespace == "aliyun" || mod.Namespace == "alibaba" {
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
			examples := make([]map[string]interface{}, 0)
			for _, example := range moduleInfo["examples"].([]interface{}) {
				examples = append(examples, example.(map[string]interface{}))
			}
			if len(examples) == 0 {
				logrus.Warningf("The Examples is empty, module name = %s, NameSpace = %s ", mod.Name, mod.Namespace)
			}
			mod.Examples = examples
			// 涉及到相关资源, 且examples不为空
			if flag && len(examples) != 0 {
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

func dependency(moduleName, examplePath, sourcePrefixName string) string {
	config := fmt.Sprintf(`
module "%s"  {
	source  = "%s//%s"
}
	`, moduleName, sourcePrefixName, examplePath)
	return config
}

func processDir(dirname, content string) error {
	err := os.MkdirAll(dirname, 0777)
	if err != nil {
		logrus.Error(err)
		return err
	}
	f, err := os.Create(dirname + "/main.tf")
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
