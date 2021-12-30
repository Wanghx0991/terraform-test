#!/bin/bash
# Created Time: 2021/11/9
# Script Description: Download module and execute
FolderName=$1

#terraform-alicloud_vpc_vpc_examples/complete
# SourceName=terraform-alicloud_vpc_vpc_examples/complete
# ModuleName= vpc
# ExamplePath = examples/complete
echo "pwd = $pwd" > ../record.log
echo  "FolderName = ${FolderName}"
cd "${FolderName}" || exit
echo "\033[33m[SKIPPED]\033[0m Current FolderName: ${FolderName}" >> ../record.log
terraform init || echo "\033[33m[Info]\033[0m terraform init Failed ${FolderName}" >> ../record.log
terraform plan || echo "\033[33m[Info]\033[0m terraform plan Failed With the With The FolderName=${FolderName}" >> ../record.log
terraform apply --auto-approve ||  echo "\033[33m[Info]\033[0m Terraform apply Failed With The FolderName = ${FolderName}*****" >> ../record.log
echo "\033[33m[Info]\033[0m Terraform apply Success With The FolderName = ${FolderName} *****" >> ../record.log
terraform destroy -force ||  echo "\033[33m[Error]\033[0m !! Terraform Destroy Failed With The FolderName = ${FolderName}*****" >> ../record.log
echo "———————————————–" >> ../record.log