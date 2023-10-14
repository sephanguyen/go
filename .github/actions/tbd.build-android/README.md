# Build Android

Build android step action

### Environment variables input

|Name|Type|Description|Require|
|--|--|--|--|
|GITHUB_TOKEN|string|github token|true|
|REPO|string|repository|true|
|RELEASE_TAG|string|get tag from [student-app](https://github.com/manabie-com/student-app/tags)|true|
|ORGANIZATION|string||true|
|ENVIRONMENT|string||true|
|APP|string|learner|true|

example:

```yaml
name: example

on:
  pull_request:
    types: [synchronize]

env:
  GITHUB_TOKEN: ${{ secrets.BUILD_GITHUB_TOKEN }}
  REPO: 'student-app'
  RELEASE_TAG: 'develop'
  ORGANIZATION: 'manabie'
  ENVIRONMENT: 'staging'
  APP: "learner"

jobs:
  example:
    runs-on: [self-hosted, Linux, docker-dev]
    steps:
      - uses: manabie-com/backend/.github/actions/build-android@develop
```
