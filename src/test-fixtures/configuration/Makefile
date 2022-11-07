plugin_dir=terraform.d/plugins/terraform.prismatic.io/prismatic-io/prismatic/0.1.0/darwin_amd64
name=terraform-provider-prismatic

build:
	cd ../.. && go build -o ${name}
	cd ../.. && mv ${name} test-fixtures/configuration/${plugin_dir}

init:
	rm -f .terraform.lock.hcl
	terraform init

plan:
	terraform plan

apply:
	terraform apply

full: build init apply

clear-state:
	rm -f terraform.tfstate terraform.tfstate.backup
