package consts

const (
	TerrafromBaseUrl       = "https://registry.terraform.io/v1/modules"
	TerraformUrl           = "https://registry.terraform.io/"
	TerraformModuleRepoUrl = "https://api.github.com/users/terraform-alicloud-modules/repos"
	PerPage                = 10
	ModulesNume            = 152
	OssEndpointBeijing     = "http://oss-cn-beijing.aliyuncs.com"
)

var NameTransfer = map[string]string{
	"api":      "api_gateway",
	"brain":    "brain_industrial_pid",
	"cen":      "cbn",
	"cdn":      "cbn",
	"click":    "click_house",
	"common":   "common_bandwith",
	"resource": "resource_manager",
	"simple":   "simple_application_server",
	"vswitch":  "vpc",
	"yundun":   "yundun_dbaoudit",
}
