#!/usr/bin/env bash

set -e

: ${ALICLOUD_ACCESS_KEY:?}
: ${ALICLOUD_SECRET_KEY:?}
: ${ALICLOUD_REGION:?}
: ${terraform_version:?}
: ${MODULE_NAME:?}
: ${DING_TALK:?}
: ${ACCESS_USER_NAME:?}
: ${ACCESS_PASSWORD:?}
: ${Module_Access_Url:?}



function PostDingTalk() {
  echo "DING_TALK= ${DING_TALK}"
  curl -X POST \
          "https://oapi.dingtalk.com/robot/send?access_token=${DING_TALK}" \
          -H 'cache-control: no-cache' \
          -H 'content-type: application/json' \
          -d "{
          \"msgtype\": \"text\",
          \"text\": {
                  \"content\": \"$1\"
          }
          }"
}
RESULT=${RESULT}"--- Terraform Module Details --- \n"
RESULT=${RESULT}"Login：${Module_Access_Url}/teams/main/pipelines/terraform-module/jobs/${MODULE_NAME} \n"
RESULT=${RESULT}"User Name：${ACCESS_USER_NAME} \n"
RESULT=${RESULT}"Password：${ACCESS_PASSWORD} \n"

export ALICLOUD_ACCESS_KEY=${ALICLOUD_ACCESS_KEY}
export ALICLOUD_SECRET_KEY=${ALICLOUD_SECRET_KEY}
export ALICLOUD_REGION=${ALICLOUD_REGION}

echo -e "\033[33m Downloading ${MODULE_NAME}\033[0m"
git clone "https://github.com/terraform-alicloud-modules/${MODULE_NAME}"

CURRENT_PATH=$(pwd)
TERRAFORM_MODULE_PATH=$CURRENT_PATH/$MODULE_NAME
EXAMPLES=$TERRAFORM_MODULE_PATH/examples
COMPLETE=$EXAMPLES/complete
TFVAR=$COMPLETE/tfvars
echo "CURRENT_PATH = $CURRENT_PATH, MODULE_NAME = ${MODULE_NAME}, TERRAFORM_MODULE_PATH = ${TERRAFORM_MODULE_PATH}, COMPLETE = ${COMPLETE}"

pushd $TERRAFORM_MODULE_PATH
if [ ! -d "${EXAMPLES}" ]; then
    RESULT="${RESULT} \n ---FAIL---\n${MODULE_NAME}: Please Add the EXAMPLES"
    echo -e "\033[33m ${RESULT} \033[0m"
    PostDingTalk "${RESULT}"
    exit 1
fi

if [ ! -d "${COMPLETE}" ]; then
    RESULT="${RESULT} \n ---FAIL---\n${MODULE_NAME}: Please Add the COMPLETE"
    echo -e "\033[33m ${RESULT} \033[0m"
    PostDingTalk "${RESULT}"
    exit 1
fi

if [ ! -d "${TFVAR}" ]; then
    RESULT="${RESULT} \n ---FAIL---\n${MODULE_NAME}: Please Add the TFVAR File"
    echo -e "\033[33m ${RESULT} \033[0m"
    PostDingTalk "${RESULT}"
    exit 1
fi

apt-get update && apt-get install -y zip
wget -qN https://releases.hashicorp.com/terraform/${terraform_version}/terraform_${terraform_version}_linux_amd64.zip
unzip -o terraform_${terraform_version}_linux_amd64.zip -d /usr/bin
if [[ "$?" == "1" ]]; then
  echo -e "\033[33m Download Terraform Error \033[0m"
  exit 1
fi
echo -e "\033[33m Current Terraform Version !\033[0m"
terraform version

pushd "$COMPLETE"
error=false
terraform init || exit 1
terraform plan || exit 1
echo -e "\033[33m Terraform Apply !\033[0m"
terraform apply --auto-approve
if [[ "$?" != "0" ]]; then
  terraform destroy --force
  RESULT="${RESULT} \n ---FAIL---\n${MODULE_NAME}: Terraform Apply Failed"
  echo -e "\033[33m ${RESULT} \033[0m"
  PostDingTalk "${RESULT}"
fi
echo -e "\033[33m Terraform Update Plan!\033[0m"
terraform plan -var-file=tfvars/01-update.tfvars
echo -e "\033[33m Terraform Update Apply!\033[0m"
terraform apply -var-file=tfvars/01-update.tfvars --auto-approve
if [[ "$?" != "0" ]]; then
  terraform destroy --force
  RESULT="${RESULT} \n ---FAIL---\n${MODULE_NAME}: Terraform Update Apply Failed"
  echo -e "\033[33m ${RESULT} \033[0m"
  PostDingTalk "${RESULT}"
fi

echo -e "\033[33m Terraform Update Plan!\033[0m"
terraform plan -var-file=tfvars/01-update.tfvars | grep "No changes. Infrastructure is up-to-date."
if [[ "$?" != "0" ]]; then
  RESULT="${RESULT} \n ---FAIL---\n${MODULE_NAME}: Terraform Update Plan Failed"
  echo -e "\033[33m ${RESULT} \033[0m"
  PostDingTalk "${RESULT}"
fi
echo -e "\033[33m Terraform Destroy!\033[0m"
terraform destroy --force || exit 1
if [[ "$?" != "0" ]]; then
  RESULT="${RESULT} \n ---FAIL---\n${MODULE_NAME}: Terraform Destroy Failed"
  echo -e "\033[33m ${RESULT} \033[0m"
  PostDingTalk "${RESULT}"
fi
RESULT="${RESULT} \n ---PASS---\n${MODULE_NAME}: Execute SUCCESS!"
PostDingTalk "${RESULT}"



