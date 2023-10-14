This is an mock repository to test git usages.

It is used by setting `GIT_DIR` environment variable. See https://git-scm.com/book/en/v2/Git-Internals-Environment-Variables

Examples:

- Getting diff

    ```sh
    GIT_DIR=internal/golibs/execwrapper/testdata/repo/test.git git diff
    ```

- Adding new files

    ```sh
    GIT_DIR=internal/golibs/execwrapper/testdata/repo/test.git g add -p
    GIT_DIR=internal/golibs/execwrapper/testdata/repo/test.git g commit -m "My commit message"
    ```

All files (`test.git` included) must still be committed and tracked by the main `.git`.
