# Get runners to run

Get runners to run


```yaml
name: example

jobs:
  runners:
    runs-on: ubuntu-latest
    outputs:
      runners: ${{ steps.runners.outputs.runners }}
    steps:
      - uses: actions/checkout@v3
        with:
          ref: develop #should be develop
      - id: runners
        uses: ./.github/actions/runners
        with:
          repo: 'backend'
          workflow: 'tbd.build'

  build:
    runs-on: ${{ fromJson(needs.runners.outputs.runners)['build-backend'] }}
    needs: runners
    outputs:
      runners: ${{ steps.runners.outputs.runners }}
    steps:
      - uses: actions/checkout@v3
      - id: runners
        uses: ./.github/actions/runners

  deploy:
    runs-on: ${{ fromJson(needs.runners.outputs.runners)['deploy-k8s'] }}
    needs: [build, runners]
    outputs:
      runners: ${{ steps.runners.outputs.runners }}
    steps:
      - uses: actions/checkout@v3
      - id: runners
        uses: ./.github/actions/runners

```
