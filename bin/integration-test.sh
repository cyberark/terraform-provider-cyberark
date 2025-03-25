#!/usr/bin/env bash
set -eo pipefail

source "$(dirname "$0")/utils.sh"

# Trim the build number from the VERSION file to be compatible with tf version constraints
export SECRETSHUB_VERSION="$(cat VERSION | cut -d'-' -f1)"

export DOCKER_COMPOSE_ARGS="-f docker-compose.test.yml"

export TF_VAR_tenant_name="${INFRAPOOL_SHARED_SERVICES_TENANT}"
export TF_VAR_domain="${INFRAPOOL_SHARED_SERVICES_DOMAIN}"
export TF_VAR_client_id="${INFRAPOOL_SHARED_SERVICES_CLIENT_ID}"
export TF_VAR_client_secret="${INFRAPOOL_SHARED_SERVICES_CLIENT_SECRET}"
export TF_VAR_aws_alias="${INFRAPOOL_SHARED_SERVICES_AWS_ALIAS}"
export TF_VAR_aws_region="${INFRAPOOL_SHARED_SERVICES_AWS_REGION}"
export TF_VAR_aws_account_id="${INFRAPOOL_SHARED_SERVICES_AWS_ACCOUNT_ID}"
export TF_VAR_aws_iam_role="${INFRAPOOL_SHARED_SERVICES_AWS_IAM_ROLE}"
export TF_VAR_source_p_cloud_id="store-77644f32-1d09-4845-91cb-574154609632"
export TF_VAR_target_secretstore_id=""

AWS_SECRETSTORE_STATEFILE="test/awssecretstore/terraform.tfstate"

function main() {
  #generate the safe name, account name, secretstore name, sync policy name
  generate_random_values

  vars_to_check=(
  "TF_VAR_tenant_name"
  "TF_VAR_domain"
  "TF_VAR_client_id"
  "TF_VAR_client_secret"
  "TF_VAR_safename"
  "TF_VAR_secret_key"
  "TF_VAR_aws_username"
  "TF_VAR_azure_username"
  "TF_VAR_db_username"
  "TF_VAR_aws_store_name"
  "TF_VAR_policy_name"
  "TF_VAR_source_p_cloud_id"
  "TF_VAR_aws_alias"
  "TF_VAR_aws_region"
  "TF_VAR_aws_account_id"
  "TF_VAR_aws_iam_role"
  "TF_VAR_aws_key_id"
  "TF_VAR_ms_app_id"
  "TF_VAR_ms_app_obj_id"
  "TF_VAR_ms_key_id"
  )

  tests_to_run=(
  "safe"
  "awsaccount"
  "azureaccount"
  "dbaccount"
  "awssecretstore"
  )

  # Perform checks
  for var in "${vars_to_check[@]}"; do
    check_var "$var"
  done
  echo ">> build image"
  dockerCompose build

  echo ">> Testing provider with tf-included vars"
  overall_status=0

  # Run and validate each test
  for test in "${tests_to_run[@]}"; do
    # Each test defines flag
    testProviderFeature "test/$test" || overall_status=1
    sleep 5
  done

  test -f "$AWS_SECRETSTORE_STATEFILE" && \
  TF_VAR_target_secretstore_id=$(cat "$AWS_SECRETSTORE_STATEFILE" | \
  jq -r '.resources[] | select(.type == "cyberark_aws_secret_store") | .instances[0].attributes.id' \
  )
  testProviderFeature "test/syncpolicy" || overall_status=1
  sleep 5
  #stop the container
  dockerCompose down

  generate_random_values
  singleRunProviderTest "test/aws_single_run" || overall_status=1

  # Summary of the results
  if [ "$overall_status" -eq 0 ]; then
    echo "######## All Terraform configurations applied and validated successfully. #######"
  else
    echo "####### One or more Terraform configurations failed. ########"
  fi

  #stop the container
  dockerCompose down

  exit $overall_status
}

function generate_random_values() {
  local random_number=$((RANDOM % 9000 + 1000))
  local random_secret_key=$(openssl rand -base64 12 | tr -dc 'A-Za-z0-9!@#$%^&*()_+[]{}|;:,.<>?')
  export TF_VAR_safename="safe_test_${random_number}"
  export TF_VAR_secret_key="$random_secret_key"
  export TF_VAR_aws_username="aws-test-${random_number}"
  export TF_VAR_azure_username="azure-test-${random_number}"
  export TF_VAR_db_username="db-test-${random_number}"
  export TF_VAR_aws_store_name="aws-store-${random_number}"
  export TF_VAR_policy_name="sync_policy_${random_number}"
  export TF_VAR_aws_key_id="$random_secret_key"
  export TF_VAR_ms_app_id=$(openssl rand -base64 12 | tr -dc 'A-Za-z0-9-')
  export TF_VAR_ms_app_obj_id=$(openssl rand -base64 12 | tr -dc 'A-Za-z0-9-')
  export TF_VAR_ms_key_id=$(openssl rand -base64 12 | tr -dc 'A-Za-z0-9-')
}

function check_var() {
  local var_name=$1
  local var_value=${!var_name}

  if [ -z "$var_value" ]; then
    echo "Error: $var_name is not set or is empty."
    return 1
  fi
}

function dockerCompose() {
  docker compose $DOCKER_COMPOSE_ARGS "$@"
}

function terraformExecute() {
  dockerCompose exec -T terraform sh -ec "$@"
}

# Function to capture and validate Terraform outputs
function validateTerraformOutputs() {
  target_dir=$1

  echo "Validating outputs for $target_dir..."

  output=$(terraformExecute \
    "cd $target_dir/ &&
     terraform output -json")

  echo "Received target_dir: $target_dir"

  value=$(cat "$target_dir/terraform.tfstate" | jq -r '.outputs.status.value')
  if [[ "$value" == "success" ]]; then
    echo "$target_dir validation was successful"
    return 0
  else
    echo "$target_dir validation was failed"
    return 1
  fi
}

function testProviderFeature() {
  target_dir=$1

  echo ">> Planning and applying '$target_dir/main.tf' Terraform manifest"

  export TF_LOG=INFO

  dockerCompose up -d terraform

  terraformExecute \
    "cd $target_dir/ &&
     terraform init &&
     terraform validate &&
     terraform plan &&
     terraform apply -auto-approve"

  # Validate outputs
  validateTerraformOutputs "$target_dir"
}

function singleRunProviderTest() {
  target_dir=$1

  echo ">> Planning and applying '$target_dir/main.tf' Terraform manifest"

  export TF_LOG=INFO

  # Start the Terraform container
  dockerCompose up -d terraform

  # Run Terraform commands
  terraform_output=$(terraformExecute \
    "cd $target_dir/ &&
     terraform init &&
     terraform validate &&
     terraform plan &&
     terraform apply -auto-approve &&
     terraform output -json")

  # Check the exit status of the `terraform apply` command
  if [ $? -ne 0 ]; then
    echo "Terraform apply failed for $target_dir."
    return 1
  fi
  # Extract the value using jq
  value=$(cat "$target_dir/terraform.tfstate" | jq -r '.outputs.overall_status.value')
  if [[ "$value" == "success" ]]; then
    echo "Single run validation was successful: $target_dir"
    return 0
  else
    echo "Single run validation failed: $target_dir"
    return 1
  fi
}

main
