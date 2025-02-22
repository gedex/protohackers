Solutions for protohackers.com
==============================

Solution for protohackers.com in various programming languages.

### Run test

Test the server.

```
make test PROBLEM=0 SERVER=go SERVER_ARGS='127.0.0.1:9999'
```

where `PROBLEM` is protohackers.com problem number, `SERVER` is the programming language
to test, and `SERVER_ARGS` is args to run the server.

### Run the server

```
make run PROBLEM=0 SERVER=zig SERVER_ARGS='127.0.0.1:9999'
```
