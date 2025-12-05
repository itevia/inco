# Welcome to inco

inco test scripts and provides the resource to the associated SAP integration flow.

# Getting started

## Install inco

```go install https://github.com/itevia/inco/cmd/inco@latest```

## Manifest usage preview
Below you can see a manifest **inco** will use as input.<br>
The manifest must be at project root.

```
tenantURL: https://<tenant>.authentication.eu10.hana.ondemand.com
apiURL: https://<tenant>.hana.ondemand.com
testPaths:
  - tools/runTests.groovy
uploadScripts:
  - id: <IFLOW ID>
    version: active
    scripts:
      - id: script1.groovy
        type: groovy
        path: src/script1.groovy
      - id: script2.groovy
        type: groovy
        path: src/script2.groovy
```
