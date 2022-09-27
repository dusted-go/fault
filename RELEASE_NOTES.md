Release Notes
=============

## 1.3.0

Changed `StackTrace()` to only return the formatted stacktrace and `String()` to return the error message with the stacktrace. Use `Error()` to only receive the error message without a stacktrace.

## 1.2.0

Added `fault.As` to the `fault` package.

## 1.1.0

Fixed issues with fault.SystemError to correctly implement the `Unwrap()` method.

## 1.0.0

Moved from dusted-go/utils.