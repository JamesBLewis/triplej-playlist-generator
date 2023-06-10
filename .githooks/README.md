# How to enable hooks

Simply run the below command in this project root:

```shell script
git config core.hooksPath .githooks
```

## List of hooks
 - pre-commit - triggers `make lint` and `make test`.