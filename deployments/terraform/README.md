# Terragrunt Infrastructure

The goal is to automate the creation of our infrastructure. Currently we are just using it for Google Cloud.

This repository contains modules defining cloud resources to be created, and a live folder that configures the modules for each environment.

### Structure

This directories contents.

```
├── live            
    # folder for terragrunt hcl files
│   ├── common.hcl
│   ├── prod-jprep
│   ├── prod-renseikai
│   ├── prod-synersia
│   ├── staging
│   │   ├── env.hcl
│   │   └── log-metrics
│   │       └── terragrunt.hcl
│   └── terragrunt.hcl
├── modules         
    # folder for re-usable modules
│   └── log-metrics
│       ├── main.tf
│       └── variables.tf
    ...
└── README.md

```


### Configuration

Edit env.hcl files in `./live/<some-folder>` to change variables we've configured.



### Example Use


```bash

export GOOGLE_APPLICATION_CREDENTIALS=/path/to/creds.json

cd live/staging/log-metrics
terragrunt plan
terragrunt apply

cd ../alert-policies
terragrunt plan
terragrunt apply

cd ../notification-channel
terragrunt plan
terragrunt apply

```

### Example 'All' Use


```bash

export GOOGLE_APPLICATION_CREDENTIALS=/path/to/creds.json

cd live/staging/
terragrunt run-all plan
terragrunt run-all apply
terragrunt run-all destroy # warning: does not prompt for yes


```