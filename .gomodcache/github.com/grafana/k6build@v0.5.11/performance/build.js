import { check } from "k6"
import http from "k6/http";
import { openKv } from "k6/x/kv";
import { AWSConfig, S3Client } from 'https://jslib.k6.io/aws/0.12.3/s3.js';

import { randomFromArray, randomFromDict } from "./lib/utils.js"

// kv tore used to coordinate builders and user
const kv = openKv();

const catalogURL  = __ENV.EXT_CATALOG_URL || 'https://registry.k6.io/product/cloud-catalog.json';
const buildSrvURL = __ENV.K6_BUILD_SERVICE_URL || 'http://localhost:8000';
const buildSrvToken = __ENV.K6_BUILD_SERVICE_AUTH || __ENV.K6_CLOUD_TOKEN
const buildSrvEndpoint = `${buildSrvURL}/build`

// set request headers. Add authorization token if defined
const buildSrvHeaders = {"Content-Type": "application/json"}
if (buildSrvToken) {
        buildSrvHeaders["Authorization"] = `Bearer ${buildSrvToken}`
}


// creates a random build request piking a combination of a k6 version and an extension
function randomBuildRequest(catalog) {
        const k6Version = randomFromArray(catalog.k6)
        const ext = randomFromDict(catalog.extensions)
        const extVersion = randomFromArray(catalog.extensions[ext])

        return {
                "k6": k6Version,
                "dependencies": [
                        { "name": ext, "constraints": "=" + extVersion }
                ],
                "platform": "linux/amd64"
        }
}

// deletes all artifacts from the store bucket
async function cleanupAWS() {
        if (!__ENV.CLEANUP_AWS || __ENV.CLEANUP_AWS == "") {
                return
        }

        const bucket = __ENV.K6_BUILD_SERVICE_BUCKET
        if ( bucket == "") {
                throw new Error("aborting AWS cleanup K6_BUILD_SERVICE_BUCKET not specified ")
        }
              
        const awsConfig = new AWSConfig({
                region: __ENV.AWS_REGION,
                accessKeyId: __ENV.AWS_ACCESS_KEY_ID,
                secretAccessKey: __ENV.AWS_SECRET_ACCESS_KEY,
                sessionToken: __ENV.AWS_SESSION_TOKEN
        });

        const s3 = new S3Client(awsConfig);

        const artifacts = await s3.listObjects(bucket);

        for (const artifact of artifacts) {
                console.log(`deleting object ${artifact.key}`)
                await s3.deleteObject(bucket, artifact.key)
        }
}

//filter extensions based on a function
// by default, filter out only k6
async function filterExtensions(catalog, filter) {
        if (!filter) {
                filter = function(entry){ return entry != "k6"}
        }

        let filtered = {}
        // collect the name and versions of all non-filtered extensions
        for (const entry of Object.keys(catalog)) {
                if (!filter(entry, catalog[entry])){
                        continue
                }
                filtered[entry] = catalog[entry].versions;
        }

        return filtered
}

export async function setup() {
        await kv.clear();

        await cleanupAWS()

        const resp = await http.get(catalogURL)
        if (resp.status != 200) {
                throw new Error(`unable to fetch catalog at ${catalogURL} : ${resp.status_text}`)
        }

        const catalog = resp.json()
        const extensions = await filterExtensions(catalog)

        return {
                "k6": catalog["k6"].versions,
                "extensions": extensions
        }
}

// make a build request
export async function build(catalog) {
        const request = JSON.stringify(randomBuildRequest(catalog))
        const resp = http.post(
                buildSrvEndpoint,
                request,
                {
                        headers: buildSrvHeaders
                }
        );

        const ok = check(resp, {
                'is status 200': (r) => r.status === 200,
                'get success': (r) => !r.json().error,
        });

        if (ok) {
                kv.set("build:" + resp.json().artifact.id, request)
        }
}

// make a request for an already-build artifact
export async function use(catalog) {
        const builds = await kv.list({ prefix: "build:" })
        if (builds.length == 0) {
                return
        }

        const request = randomFromArray(builds.map( e => e.value))

        const resp = http.post(
                buildSrvEndpoint,
                request,
                {
                        headers: buildSrvHeaders
                }
        );

        check(resp, {
                'is status 200': (r) => r.status === 200,
                'get success': (r) => !r.json().error,
        });
}

export let options = {
        scenarios: {
                builder: {
                        executor: "constant-arrival-rate",
                        rate: __ENV.BUILD_RATE || 1,
                        timeUnit: "1m",
                        duration: __ENV.TEST_DURATION || "10m" ,
                        preAllocatedVUs: 1,
                        maxVUs: 3,
                        exec: "build",
                },
                user: {
                        executor: "constant-arrival-rate",
                        rate: __ENV.USER_RATE || 10,
                        timeUnit: "1m",
                        startTime: "1m",
                        duration: __ENV.TEST_DURATION || "10m" ,
                        preAllocatedVUs: 5,
                        maxVUs: 10,
                        exec: "use",
                }
        },
}
