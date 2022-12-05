name=terraform-provider-prismatic
version=0.1.1
plugin_dir=~/.terraform.d/plugins/registry.terraform.io/prismatic-io/prismatic/${version}/darwin_amd64

build:
	cd ../../.. && go install
	mkdir -p ${plugin_dir}
	cp ${GOPATH}/bin/${name} ${plugin_dir}/${name}_${version}

init:
	rm -f .terraform.lock.hcl
	terraform init

plan:
	terraform plan

apply:
	terraform apply

destroy:
	terraform destroy

full: build init apply

clear-state:
	rm -f terraform.tfstate terraform.tfstate.backup
