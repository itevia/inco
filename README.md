# Welcome to inco

inco tests scripts and provides the resources to the associated SAP integration flow.<br/>
To be used in CICD pipelines to easily test and upload every script used in you SAP iflow.<br/>
inco will search a manifest file inco.yaml where it is invoked.

# Getting started

## Install Dependencies

Install groovy<br/>
```sudo apt-get install groovy```<br/>
Then verify installation<br/>
```groovy --version```



## Install inco

```go install https://github.com/itevia/inco/cmd/inco@latest```

or

```sudo curl -L https://github.com/itevia/inco/releases/download/0.0.1/inco-linux-amd64 -o /usr/local/bin/inco && sudo chmod +x /usr/local/bin/inco```

Welcome to the clipper wiki!

## Required Configuration

### Environment variables

| Field Name | Additional info    |
|------------|--------------------|
| CPI_USER       |  
| CPI_PASSWORD   | 


### Project configuration file - Manifest

Create a inco.yaml file at root of your project

#### Root Object

| Field Name | Type | Additional info    |
|------------|------|--------------------|
| tenantURL       |string| Required 
| apiURL      |string| Required
| testPaths       |[string]| Required - list of paths to find test scripts to run
| uploadScripts      |[Iflow]| Required - list of UploadScript

#### Iflow Object

| Field Name | Type | Additional info |
|------------|------|-----------------|
| id       |string| Required - 
| version      |string| Required - `active` or `x.x.x`
| scripts     |[Script]| Required - 

#### Script Object

| Field Name | Type | Additional info |
|------------|------|-----------------|
| id       |string| Required - 
| type      |string| Required - `groovy`
| path     |string| Required - path to find script to upload



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
