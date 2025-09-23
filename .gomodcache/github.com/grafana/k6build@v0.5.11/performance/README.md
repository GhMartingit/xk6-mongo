# k6build service load test

This folder contains a performance tests for the build service using k6.

The test runs two scenarios. In one, makes build requests for a combinations of `k6` and the extensions that (with hight probability) have not been previously built, so stress the build service.

In the other scenario, running in parallel, make requests for artifacts already built (in the build process).

## Requirements

### xk6-kv

The test use the [xk6-kv](https://github.com/oleiade/xk6-kv) extensions which provides a mechanism for sharing status between VUs in the test.

### AWS configuration

If requested (see #configuration), the test will cleanup the aws bucket used to store the artifacts. In order to do so, it will require AWS credentials and other related configuration:

* K6_BUILD_SERVICE_BUCKET: name of the aws bucket used as artifact store
* AWS_REGION: region where the bucket resided
* AWS credentials:
  - AWS_ACCESS_KEY_ID
  - AWS_SECRET_ACCESS_KEY
  - AWS_SESSION_TOKEN

## Configuration

The test can be configured using the following environment variables:

* `K6_BUILD_SERVICE_URL`: url to the k6 build service. Defaults to `http://localhost:8000`
* `K6_BUILD_SERVICE_AUTH`: token used to authenticate to the build service. If not provided, the test will default to `K6_CLOUD_TOKEN`
* `CLEANUP_AWS`: If set to a non empty value, the test will remove all objects from the aws bucket used as artifact store (defined in the [aws configuration](#aws-configuration)) during the setup
* `TEST_DURATION`: duration of the test. Defaults to `10m`
* `BUILD_RATE`: rate in of build request per minute. Defaults to `10`
* `USER_RATE`: rate of request per minute of build requests for artifacts already built. Defaults to `10`