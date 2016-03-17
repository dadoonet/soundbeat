# Soundbeat

Welcome to Soundbeat.

Ensure that this folder is at the following location:
`${GOPATH}/github.com/dadoonet`

## Getting Started with Soundbeat

### Init Project
To get running with Soundbeat, run the following commands:

```
make init
```


To push Soundbeat in the git repository, run the following commands:

```
git commit 
git remote set-url origin https://github.com/dadoonet/soundbeat
git push origin master
```

For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).

### Build

To build the binary for Soundbeat run the command below. This will generate a binary
in the same directory with the name soundbeat.

```
make
```


### Run

To run Soundbeat with debugging output enabled, run:

```
./soundbeat -c soundbeat.yml -e -d "*"
```


### Test

To test Soundbeat, run the following commands:

```
make testsuite
```

alternatively:
```
make unit-tests
make system-tests
make integration-tests
make coverage-report
```

The test coverage is reported in the folder `./build/coverage/`


### Package

To cross-compile and package Soundbeat for all supported platforms, run the following commands:

```
cd dev-tools/packer
make deps
make images
make
```

### Update

Each beat has a template for the mapping in elasticsearch and a documentation for the fields
which is automatically generated based on `etc/fields.yml`.
To generate etc/soundbeat.template.json and etc/soundbeat.asciidoc

```
make update
```


### Cleanup

To clean  Soundbeat source code, run the following commands:

```
make fmt
make simplify
```

To clean up the build directory and generated artifacts, run:

```
make clean
```


### Clone

To clone Soundbeat from the git repository, run the following commands:

```
mkdir -p ${GOPATH}/github.com/dadoonet
cd ${GOPATH}/github.com/dadoonet
git clone https://github.com/dadoonet/soundbeat
```


For further development, check out the [beat developer guide](https://www.elastic.co/guide/en/beats/libbeat/current/new-beat.html).
