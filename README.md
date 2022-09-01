# Fault

Custom Go errors to allow a richer error handling experience in web applications.

It provides two custom `error` structs:

- `fault.System`
- `fault.User`

System faults are errors which are normally not caused by improper use of a library or function. Those are errors which in a web application would typically get logged and result in a 5xx error response since the user cannot do anything about it (e.g. service outage, DNS issues, IO errors when reading/writing something, etc.)

User faults are errors which can be avoided by the end user. Those errors are normally returned to an end user in order to explain to them how to fix the issue (e.g. providing a wrong secret, requiring authentication, invoking an API with invalid parameters or calling a resource which does not exist).

The `fault` package allows domain code to return one of these two error types so that higher level application code (e.g. a CLI app or web API) can then decide how to correctly deal with an `error` coming from a domain layer.
