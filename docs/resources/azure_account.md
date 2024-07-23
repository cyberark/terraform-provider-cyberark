---
# generated by https://github.com/hashicorp/terraform-plugin-docs
page_title: "cybr-sh_azure_account Resource - cybr-sh"
subcategory: ""
description: |-
  Microsoft Azure Account Resource
  This resource is responsible for creating a new privileged account that contains all the required Azure information as mentioned below in Privilege Cloud.
  For more information click here https://docs.cyberark.com/privilege-cloud-shared-services/latest/en/Content/WebServices/Add%20Account%20v10.htm.
---

# cybr-sh_azure_account (Resource)

Microsoft Azure Account Resource

This resource is responsible for creating a new privileged account that contains all the required Azure information as mentioned below in Privilege Cloud.

For more information click [here](https://docs.cyberark.com/privilege-cloud-shared-services/latest/en/Content/WebServices/Add%20Account%20v10.htm).

## Example Usage

```terraform
variable "secret_key" {
  type      = string
  sensitive = true
}

resource "cybr-sh_azure_account" "mskey" {
  name             = "user-ms"
  address          = "1.2.3.4"
  username         = "user-ms"
  platform         = "MS_TF"
  safe             = "TF_TEST_SAFE"
  secret           = var.secret_key
  sm_manage        = false
  sm_manage_reason = "No CPM Associated with Safe."
  ms_app_id        = "Application ID"
  ms_app_obj_id    = "Application Object ID"
  ms_key_id        = "Key ID"
  ms_ad_id         = "AD Key ID"
  ms_duration      = "300"
  ms_pop           = "yes"
  ms_key_desc      = "key descriptiong with spaces"
}
```

<!-- schema generated by tfplugindocs -->
## Schema

### Required

- `address` (String) URI, URL or IP associated with the credential.
- `ms_app_id` (String) Microsoft Azure Application ID.
- `ms_app_obj_id` (String) Microsoft Azure Application Object ID.
- `ms_key_id` (String) Microsoft Azure Key ID.
- `name` (String) Custom Account Name for customizing the object name in a safe.
- `platform` (String) Management Platform associated with the Database Credential.
- `safe` (String) Target Safe where the credential object will be onboarded.
- `secret` (String, Sensitive) Password of the credential object.
- `username` (String) Username of the Credential object.

### Optional

- `ms_ad_id` (String) Microsoft Azure Active Directory ID.
- `ms_duration` (String) Duration.
- `ms_key_desc` (String) Key Description.
- `ms_pop` (String) Populate if not exist.
- `sm_manage` (Boolean) Automatic Management of a credential. Optional Value.
- `sm_manage_reason` (String) If sm_manage is false, provide reason why credential is not managed.

### Read-Only

- `id` (String) CyberArk Privilege Cloud Credential ID- Generated from CyberArk after onboarding account into a safe.
- `last_updated` (String)
- `secret_type` (String) Should always be 'password' for Azure Account.