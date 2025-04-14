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

function main() {
  generate_random_values

  vars_to_check=(
  "TF_VAR_tenant_name"
  "TF_VAR_domain"
  "TF_VAR_client_id"
  "TF_VAR_client_secret"
  "TF_VAR_safe_name"
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

  # Perform checks
  for var in "${vars_to_check[@]}"; do
    check_var "$var"
  done

  echo ">> build image"
  dockerCompose build

  echo ">> Testing provider with tf-included vars"

  declare -a test_results=()

  # Run CRUD tests for each resource and collect results

  echo ">> Running Safe CRUD test"
  test_resource_crud "test/safe" "safe" "safe_name"
  test_results+=("Safe|$?")

  echo ">> Running AWS Account CRUD test"
  test_resource_crud "test/aws_account" "aws_account" "account_id"
  test_results+=("AWS Account|$?")

  echo ">> Running Azure Account CRUD test"
  test_resource_crud "test/azure_account" "azure_account" "account_id"
  test_results+=("Azure Account|$?")

  echo ">> Running DB Account CRUD test"
  test_resource_crud "test/db_account" "db_account" "account_id"
  test_results+=("DB Account|$?")

  echo ">> Running AWS Secret Store CRUD test"
  test_resource_crud "test/aws_secret_store" "aws_secret_store" "store_id"
  test_results+=("AWS Secret Store|$?")

  echo ">> Running Sync Policy CRUD test"
  test_resource_crud "test/sync_policy" "sync_policy" "policy_id"
  test_results+=("Sync Policy|$?")

  # Stop the container
  dockerCompose down

  # Report results and set exit code
  report_test_results "${test_results[@]}"
  exit $?
}

function generate_random_values() {
  local random_number=$((RANDOM % 9000 + 1000))
  local random_secret_key=$(openssl rand -base64 12 | tr -dc 'A-Za-z0-9!@#$%^&*()_+[]{}|;:,.<>?')
  export TF_VAR_safe_name="safe_test_${random_number}"
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

function test_resource_crud() {
  local test_dir="$1"
  local resource_name="$2"
  local id_output_name="$3" # Output variable name for the resource ID

  echo "Testing ${resource_name} CRUD operations"

  # Generate random values for the test
  generate_random_values

  dockerCompose up -d terraform

  # Cleanup any previous state
  cleanup_terraform "$test_dir"

  # Step 1: Create Resource
  echo "Step 1: Creating ${resource_name}"
  if ! terraformExecute "cd $test_dir/create && \
    terraform init && \
    terraform validate && \
    terraform plan && \
    terraform apply -auto-approve"; then
    echo "FAILED: ${resource_name} creation failed"
    return 1
  fi

  # Get the resource ID for validation
  local resource_id=$(terraformExecute "cd $test_dir/create && jq -r '.outputs.${id_output_name}.value' terraform.tfstate")
  if [[ -z "$resource_id" || "$resource_id" == "null" ]]; then
    echo "FAILED: Could not get $id_output_name from terraform output"
    return 1
  fi
  echo "Created ${resource_name} with ID: $resource_id"

  # Step 2: Import the Resource
  echo "Step 2: Importing ${resource_name}"
  if ! terraformExecute "cd $test_dir/update && \
    terraform init && \
    terraform import cyberark_${resource_name}.imported $resource_id"; then
    echo "FAILED: ${resource_name} import command failed"
    return 1
  fi

  # Verify import was successful
  local import_resource_id=$(terraformExecute "cd $test_dir/update && jq -r '.resources[] | select(.type == \"cyberark_${resource_name}\") | .instances[0].attributes.id' terraform.tfstate")
  if [[ "$import_resource_id" != "$resource_id" ]]; then
    echo "FAILED: ${resource_name} import verification failed"
    return 1
  fi
  echo "${resource_name} import successful"

  # Step 3: Update the Resource
  echo "Step 3: Updating ${resource_name}"
  if ! terraformExecute "cd $test_dir/update && \
    terraform validate && \
    terraform plan && \
    terraform apply -auto-approve"; then
    echo "FAILED: ${resource_name} update command failed"
    return 1
  fi

  # Verify update was successful
  local update_status=$(terraformExecute "cd $test_dir/update && jq -r '.outputs.update_status.value // \"fail\"' terraform.tfstate")
  if [[ "$update_status" != "success" ]]; then
    echo "FAILED: ${resource_name} update verification failed with status: $update_status"
    return 1
  fi

  # Step 4: Delete the Resource
  echo "Step 4: Deleting ${resource_name}"
  if ! terraformExecute "cd $test_dir/update && terraform destroy -auto-approve"; then
    echo "FAILED: ${resource_name} deletion failed"
    return 1
  fi

  echo "SUCCESS: ${resource_name} CRUD testing completed successfully"
  return 0
}

function cleanup_terraform() {
  local test_dir="$1"
  terraformExecute "cd $test_dir/create && rm -rf .terraform* terraform.tfstate*"
  terraformExecute "cd $test_dir/update && rm -rf .terraform* terraform.tfstate*"
}

function report_test_results() {
  echo "=========================================="
  echo "         TEST RESULTS SUMMARY             "
  echo "=========================================="

  local results=("$@")
  local total_tests=${#results[@]}
  local passed_tests=0
  local failed_tests=0

  # Print each test result
  for result in "${results[@]}"; do
    IFS='|' read -r name status <<< "$result"
    if [[ "$status" -eq 0 ]]; then
      echo "✅ $name: PASSED"
      passed_tests=$((passed_tests+1))
    else
      echo "❌ $name: FAILED"
      failed_tests=$((failed_tests+1))
    fi
  done

  echo "=========================================="
  echo "SUMMARY: $passed_tests of $total_tests tests passed, $failed_tests failed"
  echo "=========================================="

  if [[ "$passed_tests" -eq "$total_tests" ]]; then
    return 0
  else
    return 1
  fi
}

main
