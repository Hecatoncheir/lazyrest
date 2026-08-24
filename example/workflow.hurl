# Hurl executes every exchange in this file as one session.
POST https://httpbin.org/anything
Accept: application/json
Content-Type: application/json
{
  "resource": "users"
}

HTTP 200
[Captures]
resource: jsonpath "$.json.resource"
[Asserts]
jsonpath "$.json.resource" == "users"

GET https://httpbin.org/anything/{{resource}}
Accept: application/json

HTTP 200
[Asserts]
jsonpath "$.url" contains "/users"
