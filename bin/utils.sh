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

function getPolicies(){
  local domain=$1 token=$2

  curl -s -H "Authorization: Bearer $token" \
    -X GET "https://$domain.secretshub.cyberark.cloud/api/policies" \
    -H 'Accept: application/json' | jq -r '.policies'
}

function deletePolicy() {
  local domain=$1 token=$2 policy_id=$3
  local status

  # Disable the policy
  status=$(curl -s -o /dev/null -w "%{http_code}" -H "Authorization: Bearer $token" \
    -X PUT "https://$domain.secretshub.cyberark.cloud/api/policies/$policy_id/state" \
    -H 'Accept: application/json' \
    -H 'Content-Type: application/json' \
    -d '{ "action": "disable" }')

  if [ "$status" -ne 200 ]; then
    echo "Failed to disable policy: $policy_id"
    echo "Trying to delete the policy anyway..."
  fi

  echo "Deleting policy: $policy_id"
  curl -s -H "Authorization: Bearer $token" \
    -X DELETE "https://$domain.secretshub.cyberark.cloud/api/policies/$policy_id" \
    -H 'Accept: application/json'
  echo " - Policy deleted: $policy_id"
}

function getSecretStores(){
  local domain=$1 token=$2

  # Exclude default PAM_PCLOUD secret store
  curl -s -H "Authorization: Bearer $token" \
    -X GET "https://$domain.secretshub.cyberark.cloud/api/secret-stores" \
    -H 'Accept: application/json' | jq -r '.secretStores | map(select(.type != "PAM_PCLOUD"))'
}

function deleteSecretStore(){
  local domain=$1 token=$2 store_id=$3

  echo "Deleting secret store: $store_id"
  curl -s -H "Authorization: Bearer $token" \
    -X DELETE "https://$domain.secretshub.cyberark.cloud/api/secret-stores/$store_id" \
    -H 'Accept: application/json'
  echo " - Secret store deleted: $store_id"
}

function getAccounts(){
  local domain=$1 token=$2

  # Exclude default accounts
  curl -s -H "Authorization: Bearer $token" \
    -X GET "https://$domain.privilegecloud.cyberark.cloud/passwordvault/api/accounts" \
    -H 'Content-Type: application/json'| jq -r '.value | map(select(.platformId != ""))'
}

function deleteAccount(){
  local domain=$1 token=$2 account_id=$3

  echo "Deleting account: $account_id"
  curl -s -H "Authorization: Bearer $token" \
    -X DELETE "https://$domain.privilegecloud.cyberark.cloud/passwordvault/api/accounts/$account_id" \
    -H 'Content-Type: application/json'
  echo " - Account deleted: $account_id"
}

function getSafes(){
  local domain=$1 token=$2

  curl -s -H "Authorization: Bearer $token" \
    -X GET "https://$domain.privilegecloud.cyberark.cloud/passwordvault/api/safes" \
    -H 'Content-Type: application/json' | jq -r '.value | map(select(.creator.name != "Administrator"))'
}

function deleteSafe(){
  local domain=$1 token=$2 safe_name=$3

  echo "Deleting safe: $safe_name"
  curl -s -H "Authorization: Bearer $token" \
    -X DELETE "https://$domain.privilegecloud.cyberark.cloud/passwordvault/api/safes/$safe_name" \
    -H 'Content-Type: application/json'
  echo " - Safe deleted: $safe_name"
}
