# Changelog

All notable changes to the Netskope Terraform Provider will be documented in this file.

The format is based on Keep a Changelog, and this project adheres to Semantic Versioning.

## Forward Statement:
The prior version base: 2.x will no longer be supported. And may not be directly compatible with future versions.
Starting with version 3.2 all new terraform providers will be built on this code base and be supported by Netskope.
This shift is to assist in terraform adoption, simplify configuration and overcome some API limitations.

## Key Features:
- Release Terraform Provider for every CRUD endpoint that is used to manage the Netskope Tenant
    - first release is focused on Netskope Private Access (NPA) v0.3.2
- Logical Attribute names 
    - instead of just "id" as an attribute, it is descriptive, "private_app_id" or "publisher_id"
- Duplicate / not used attributes are removed from the resource, and tfstate
- Mask Confusing api input and response
    - before:
        - input: private_app_name: "my app" response: private_app_name: "[my app]"
    - after: (starting v.0.3.2)
        - input: private_app_name: "my app" response: private_app_name: "my app" 
        - Terraform provider will address api input/output and normalize to maintain plan and state sanity.


[v.0.3.2] - First Release

Added
- Netskope Private Access
- Netskope Private Access Publisher Support
    - Creation of Registration Tokens
    - Management of Upgrade Profiles
    - Management of availble versions
    - Management of Publisher Alert Profiles
- Netskope Private Access Policy Groups
- Netskope Private Access Policy Rules 

Changed
 - N/A first release of new code base

Deprecated
- N/A

Removed
- v.0.2.6 Is no longer supported.


