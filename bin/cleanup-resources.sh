#!/usr/bin/env bash
set -eo pipefail

source "$(dirname "$0")/utils.sh"

# Check required environment variables
required_vars=(
  "INFRAPOOL_SHARED_SERVICES_TENANT"
  "INFRAPOOL_SHARED_SERVICES_DOMAIN"
  "INFRAPOOL_SHARED_SERVICES_CLIENT_ID"
  "INFRAPOOL_SHARED_SERVICES_CLIENT_SECRET"
)
missing_vars=()

for var in "${required_vars[@]}"; do
  if [ -z "${!var}" ]; then
    missing_vars+=("$var")
  fi
done

if [ ${#missing_vars[@]} -gt 0 ]; then
  echo "Error: The following required environment variables are not set:"
  for var in "${missing_vars[@]}"; do
    echo "  - $var"
  done
  echo "Please set these variables and try again."
  exit 1
fi

main() {
  echo "Starting cleanup of resources..."

  # Generate authentication token
  echo "Generating authentication token..."
  token=$(generateToken "$INFRAPOOL_SHARED_SERVICES_TENANT" "$INFRAPOOL_SHARED_SERVICES_CLIENT_ID" "$INFRAPOOL_SHARED_SERVICES_CLIENT_SECRET")
  if [ -z "$token" ]; then
    echo "Failed to generate authentication token. Exiting."
    exit 1
  fi
  echo "Token generated successfully."

  # Delete all test sync policies
  echo "Deleting all test sync policies from SecretsHub..."
  policies=$(getPolicies "$INFRAPOOL_SHARED_SERVICES_DOMAIN" "$token")

  # Iterate over policies JSON array
  echo "$policies" | jq -c '.[]' | while read -r policy; do
    policy_id=$(echo "$policy" | jq -r '.id')
    echo "Deleting policy: $policy_id"
    deletePolicy "$INFRAPOOL_SHARED_SERVICES_DOMAIN" "$token" "$policy_id"
  done

  # Delete all test secret stores
  echo "Deleting all test secret stores from SecretsHub..."
  secret_stores=$(getSecretStores "$INFRAPOOL_SHARED_SERVICES_DOMAIN" "$token")

  # Iterate over secret stores JSON array
  echo "$secret_stores" | jq -c '.[]' | while read -r store; do
    store_id=$(echo "$store" | jq -r '.id')
    deleteSecretStore "$INFRAPOOL_SHARED_SERVICES_DOMAIN" "$token" "$store_id"
  done

  # Delete all test accounts from pCloud
  echo "Deleting all test accounts from pCloud..."
  accounts=$(getAccounts "$INFRAPOOL_SHARED_SERVICES_DOMAIN" "$token")

  # Iterate over accounts JSON array
  echo "$accounts" | jq -c '.[]' | while read -r account; do
    account_id=$(echo "$account" | jq -r '.id')
    deleteAccount "$INFRAPOOL_SHARED_SERVICES_DOMAIN" "$token" "$account_id"
  done

  # Delete all test safes from pCloud - do this after accounts
  echo "Deleting all test safes from pCloud..."
  safes=$(getSafes "$INFRAPOOL_SHARED_SERVICES_DOMAIN" "$token")

  # Iterate over safes JSON array
  echo "$safes" | jq -c '.[]' | while read -r safe; do
    safe_name=$(echo "$safe" | jq -r '.safeName')
    deleteSafe "$INFRAPOOL_SHARED_SERVICES_DOMAIN" "$token" "$safe_name"
  done

  echo "Cleanup completed successfully."
}

main

