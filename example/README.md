# Examples

Run lazyrest against this directory from the repository root:

```sh
go run . example
```

The `.http` files demonstrate named requests, JSON/XML/GraphQL bodies,
recursive variables, and a request that uses what an earlier one answered. The `.hurl` files demonstrate response assertions and a
multi-request workflow with a captured value. Hurl examples require the
`hurl` executable.

The requests use the public `https://httpbin.org` test service and therefore
require internet access. They contain no real credentials or destructive
operations.
