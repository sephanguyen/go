# Get latest tag on trunk
Get latest tag for BE, ME, FE on trunk(develop)

### Inputs
|Name|Type|Description|Require|
|--|--|--|--|
|github-token|string|github token|true|

### Outputs
|Name|Type|Description|
|--|--|--|
|fe_tag|string|FE latest tag on trunk|
|me_tag|string|ME latest tag on trunk|
|be_tag|string|BE latest tag on trunk|


```yaml
name: example

on:
  # type

env:
  GITHUB_TOKEN: ${{ secrets.BUILD_GITHUB_TOKEN }}

jobs:
  example:
    runs-on: ubuntu-latest
    steps:
      - name: Get latest tag on trunk
        id: latest-tag
        uses: manabie-com/backend/.github/actions/tbd.get-latest-tag-on-trunk@develop
        with:
          github-token: ${{ secrets.BUILD_GITHUB_TOKEN }}
      - name: Print latest tag
        shell: bash
        run: |
          echo FE_TAG is ${{ steps.latest-tag.outputs.fe_tag }}
          echo ME_TAG is ${{ steps.latest-tag.outputs.me_tag }}
          echo BE_TAG is ${{ steps.latest-tag.outputs.be_tag }}
```