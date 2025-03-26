#!/bin/bash -x

source "$(dirname "$0")/utils.sh"

function generate_random_values() {
  local random_number=$((RANDOM % 9000 + 1000))
  local random_secret_key=$(openssl rand -base64 12 | tr -dc 'A-Za-z0-9!@#$%^&*()_+[]{}|;:,.<>?')
  export TF_SAFE_NAME="safe_acceptance_test_${random_number}"
  export TF_AWS_NAME="aws-acceptance-test-${random_number}"
  export TF_AWS_USERNAME="$TF_AWS_NAME"
  export TF_AWS_KEY_ID="$random_secret_key"
  export TF_AWS_SECRET="$random_secret_key"
  export TF_AZURE_NAME="azure-acceptance-test-${random_number}"
  export TF_AZURE_USERNAME="$TF_AZURE_NAME"
  export TF_AZURE_SECRET="$random_secret_key"
  export TF_DB_NAME="db-acceptance-test-${random_number}"
  export TF_DB_USERNAME="$TF_DB_NAME"
  export TF_DB_SECRET="$random_secret_key"
  export TF_AZURE_APP_ID=$(openssl rand -base64 12 | tr -dc 'A-Za-z0-9-')
  export TF_AZURE_OBJ_ID=$(openssl rand -base64 12 | tr -dc 'A-Za-z0-9-')
  export TF_AZURE_KEY_ID=$(openssl rand -base64 12 | tr -dc 'A-Za-z0-9-')
}

function main() {
  generate_random_values

  export PATH="$(pwd):$PATH"
  echo "Path: $PATH"

  echo "Running go tests"
  echo "Current dir: $(pwd)"

  mkdir -p output

  echo "Running unit tests..."
  go test --coverprofile=output/unit-c.out -v ./internal/cyberark | tee output/unit-junit.output

  # Convert verbose test output to JUnit XML format
  echo "Converting unit test output to JUnit XML..."
  go-junit-report < output/unit-junit.output > output/unit-junit.xml

  # Run acceptance tests, generate coverage profile, and verbose test output
  echo "Running acceptance tests..."
  TF_ACC=1 go test --coverprofile=output/acceptance-c.out -v ./internal/provider | tee output/acceptance-junit.output

  # Convert verbose test output to JUnit XML format
  echo "Converting acceptance test output to JUnit XML..."
  go-junit-report < output/acceptance-junit.output > output/acceptance-junit.xml

  # Merge coverage profiles using gocovmerge
  echo "Merging coverage profiles..."
  gocovmerge output/unit-c.out output/acceptance-c.out > output/combined-c.out
  echo "Coverage profile merged into output/combined-c.out."

  # Convert the merged coverage profile to XML format
  echo "Converting merged coverage profile to XML format..."
  gocov convert output/combined-c.out | gocov-xml > output/coverage.xml
  echo "Coverage report generated at output/coverage.xml."

  # Combine verbose test outputs and convert to JUnit XML format
  echo "Combining verbose test outputs and converting to JUnit XML..."
  cat output/unit-junit.output output/acceptance-junit.output > output/junit.output
  go-junit-report < output/junit.output > output/junit.xml
  rm output/junit.output output/acceptance* output/unit* output/combined-c.out
}

main
