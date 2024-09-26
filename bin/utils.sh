#!/bin/bash

# Function to generate OIDC token
function generateToken() {
  local tenant=$1 client_id=$2 client_secret=$3
  local token

  token=$(curl -s --location --request POST "https://$tenant.id.cyberark.cloud/oauth2/platformtoken" \
    --header 'Content-Type: application/x-www-form-urlencoded' \
    --data-urlencode "grant_type=client_credentials" \
    --data-urlencode "client_id=$client_id" \
    --data-urlencode "client_secret=$client_secret" | jq -r '.access_token')

  if [ -z "$token" ]; then
    echo "Failed to obtain access token."
    exit 1
  fi
  echo "$token"
}

# Function to disable and delete policy
function disableAndDeletePolicy() {
  local domain=$1 token=$2 policy_id=$3
  local status
  
  # Disable the policy
  status=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $token" \
    -X PUT "https://$domain.secretshub.cyberark.cloud/api/policies/$policy_id/state" \
    -H 'Accept: application/json' \
    -H 'Content-Type: application/json' \
    -d '{ "action": "disable" }')
  
  if [ "$status" -eq 200 ]; then
    # Delete the policy
    status=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $token" \
      -X DELETE "https://$domain.secretshub.cyberark.cloud/api/policies/$policy_id" \
      -H 'Accept: application/json')

    [ "$status" -eq 200 ] && echo "Successfully deleted the sync policy: $policy_id" || echo "Failed to delete the sync policy($policy_id): $status"
  else
    echo "Failed to disable the sync policy: $policy_id"
  fi
}

# Function to delete secret store
function deleteSecretStore() {
  local domain=$1 token=$2 store_id=$3
  local status

  # Delete the secret store
  status=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $token" \
    -X DELETE "https://$domain.secretshub.cyberark.cloud/api/secret-stores/$store_id" \
    -H 'Accept: */*')

  if [ "$status" -eq 204 ]; then
    echo "Successfully deleted the SecretStore: $store_id"
  else
    echo "Failed to delete the SecretStore($store_id): $status"
  fi
}

# Function to get policy ID from terraform.tfstate
function getPolicyID() {
  local dir=$1
  if [ -d "$dir" ]; then
    jq -r '.resources[] | select(.type == "cyberark_sync_policy") | .instances[0].attributes.id' \
      < "$dir/terraform.tfstate"
  fi
}

# Function to get store ID From API
function getStoreID(){
   local domain=$1 token=$2
    curl -s -H "Authorization: Bearer $token" \
    -X GET "https://$domain.secretshub.cyberark.cloud/api/secret-stores?behavior=SECRETS_TARGET&filter=type%20EQ%20AWS_ASM" \
    -H 'Accept: application/json' | jq -r '.secretStores[] | select(.name == "aws_store") | .id'
}

# Function to fetch Policy ID from API
function fetchPolicyIDFromApi(){
  local domain=$1 token=$2 store_id=$3
   if [ -n "$store_id" ]; then
     curl -s -H "Authorization: Bearer $token" \
      -X GET "https://$domain.secretshub.cyberark.cloud/api/policies?filter=target.id%20EQ%20$store_id" \
      -H 'Accept: application/json' | jq -r '.policies[].id' 
   fi
}
