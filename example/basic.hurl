# A single Hurl exchange with response assertions.
GET https://httpbin.org/get?source=lazyrest
Accept: application/json
X-Client: lazyrest

HTTP 200
[Asserts]
header "content-type" contains "application/json"
jsonpath "$.args.source" == "lazyrest"
jsonpath "$.headers.X-Client" == "lazyrest"
