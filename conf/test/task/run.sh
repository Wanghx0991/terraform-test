#!/usr/bin/env bash

set -e

: ${ALICLOUD_ACCESS_KEY:?}
: ${ALICLOUD_SECRET_KEY:?}
: ${ALICLOUD_ACCESS_KEY_SLAVE:?}
: ${ALICLOUD_SECRET_KEY_SLAVE:?}
: ${USE_SLAVE:?}
: ${ALICLOUD_REGION:?}
: ${terraform_version:?}
: ${MODULE_NAME:?}
: ${DING_TALK:?}
: ${ACCESS_USER_NAME:?}
: ${ACCESS_PASSWORD:?}
: ${Module_Access_Url:?}
: ${Update:?}

PINK='\E[1;35m'        #粉红
RES='\E[0m'

function PostDingTalk() {
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

if [[ $USE_SLAVE == "true" ]]; then
  export ALICLOUD_ACCESS_KEY=${ALICLOUD_ACCESS_KEY_SLAVE}
  export ALICLOUD_SECRET_KEY=${ALICLOUD_SECRET_KEY_SLAVE}
  export ALICLOUD_REGION=${ALICLOUD_REGION}
else
  export ALICLOUD_ACCESS_KEY=${ALICLOUD_ACCESS_KEY}
  export ALICLOUD_SECRET_KEY=${ALICLOUD_SECRET_KEY}
  export ALICLOUD_REGION=${ALICLOUD_REGION}
fi


echo -e  "${PINK}======Current AK ${ALICLOUD_ACCESS_KEY}======${RES}"

echo -e  "${PINK}======Current Region ${ALICLOUD_REGION}======${RES}"
echo -e  "${PINK}======Whether to Update:  ${Update}======${RES}"

echo -e  "${PINK}======Downloading ${MODULE_NAME}======${RES}"
rm -rf ${MODULE_NAME}

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

if [ ! -d "${TFVAR}" ] && [ Update == "true" ] ; then
    RESULT="${RESULT} \n ---FAIL---\n${MODULE_NAME}: Please Add the TFVAR File"
    echo -e "\033[33m ${RESULT} \033[0m"
    PostDingTalk "${RESULT}"
    exit 1
fi

apt-get update && apt-get install -y zip
wget -qN https://releases.hashicorp.com/terraform/${terraform_version}/terraform_${terraform_version}_linux_amd64.zip
unzip -o terraform_${terraform_version}_linux_amd64.zip -d /usr/bin


echo -e  "${PINK}====== Current Terraform Version ! ======${RES}"
terraform version

pushd "$COMPLETE"

terraform init
terraform plan || {
    RESULT="${RESULT} \n ---FAIL---\n${MODULE_NAME}: Terraform Plan Failed"
    echo -e "\033[33m ${RESULT} \033[0m"
    PostDingTalk "${RESULT}"
    exit 1
}


echo -e  "${PINK}====== Terraform Apply ======${RES}"
terraform apply --auto-approve || {
    RESULT="${RESULT} \n ---FAIL---\n${MODULE_NAME}: Terraform Apply Failed"
    echo -e "\033[33m ${RESULT} \033[0m"
    PostDingTalk "${RESULT}"
    terraform destroy -auto-approve
    exit 1
}



if [ $Update == "true" ]; then
    echo -e  "${PINK}====== Terraform Update Plan! ======${RES}"
    terraform plan -var-file=tfvars/01-update.tfvars || {
        RESULT="${RESULT} \n ---FAIL---\n${MODULE_NAME}: Terraform Plan -var-file=tfvars/01-update.tfvars Failed"
        echo -e "\033[33m ${RESULT} \033[0m"
        PostDingTalk "${RESULT}"
        terraform destroy -auto-approve
        exit 1
    }


    echo -e  "${PINK}====== Terraform Update Apply! ======${RES}"
    terraform apply -var-file=tfvars/01-update.tfvars --auto-approve || {
        RESULT="${RESULT} \n ---FAIL---\n${MODULE_NAME}: Terraform Update Apply Failed"
        echo -e "\033[33m ${RESULT} \033[0m"
        PostDingTalk "${RESULT}"
        terraform destroy -auto-approve
        exit 1
    }


    echo -e  "${PINK}====== Terraform Update Plan! ======${RES}"
    sleep 3

    Failed=true
    lines=$(terraform plan -var-file=tfvars/01-update.tfvars -lock=false -no-color)
    while read line; do
      echo $line
      if [[ "${line}" == "No changes. Infrastructure is up-to-date."* || "${line}" == "Plan: 0 to add, 0 to change, 0 to destroy."* ]];then
        Failed=false
      fi
    done <<< "$lines"
    if $Failed ; then
        RESULT="${RESULT} \n ---FAIL---\n${MODULE_NAME}: Terraform Update Plan Check Failed"
        echo -e "\033[33m ${RESULT} \033[0m"
        PostDingTalk "${RESULT}"
        terraform destroy -auto-approve
        exit 1
    fi
    echo -e  "${PINK}====== Terraform Update Plan Check Success !!======${RES}"
fi


echo -e  "${PINK}====== Wait For The Status Sync! ======${RES}"
sleep 15
echo -e  "${PINK}====== Terraform Destroy! ======${RES}"
terraform destroy -auto-approve
if [ $? -ne 0 ]; then
  RESULT="${RESULT} \n ---FAIL---\n${MODULE_NAME}: Terraform Destroy Failed"
  echo -e "\033[33m ${RESULT} \033[0m"
  PostDingTalk "${RESULT}"
  terraform destroy -auto-approve
fi

# Success -> Send the result to the robot
RESULT="${RESULT} \n ---PASS---\n${MODULE_NAME}: Execute SUCCESS!"
PostDingTalk "${RESULT}"


