### **Required:**
- Salesforce CLI: https://developer.salesforce.com/docs/atlas.en-us.sfdx_setup.meta/sfdx_setup/sfdx_setup_install_cli.htm
- NodeJS: https://nodejs.org/en/download/ (TODO: should remove this dependency)

### **Retrieve Package:**
```
scripts/salesforce/retrieve_package.bash example-module dev_hub
```
 example-module: the name of the package in folder `scripts/salesforce/salesforce`
 dev_hub: `Alias` in you org. To get the `Alias` run command `sf org list`

 ```
 Non-scratch orgs
================================================================================================================
|   ALIAS      USERNAME                    ORG ID             CONNECTED STATUS
| ─ ────────── ─────────────────────────── ────────────────── ──────────────────────────────────────────────────
|   dev_hub    minhthao.nguyen@manabie.com 00D5j00000BcJq8EAF Connected
 ```

To see detail of the org run command `sf org:display -o dev_hub`

 ```
  KEY              VALUE
 ──────────────── ────────────────────────────────────────────────────────────────────────────────────────────────────────────────
 Access Token     00D5j00000BcJq8!ARgAQDol7uQrkCAz9MeoSy9urFntWg_fNvckZ79g0IjeH0FibVk2SX6y00892tqr9NdmgfjMqbBaT7jBntTHSPzwTF2FszJq
 Alias            dev_hub
 Api Version      58.0
 Client Id        PlatformCLI
 Connected Status Connected
 Id               00D5j00000BcJq8EAF
 Instance Url     https://manabie4-dev-ed.develop.my.salesforce.com
 Username         minhthao.nguyen@manabie.com
 ```