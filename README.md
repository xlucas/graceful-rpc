# graceful-rpc

A fork of the `net/rpc` package that implements graceful shutdown.

## Why ?

The `net/rpc` package has some interesting features, it notably doesn't rely on
an IDL and is really fast. But it lacks a graceful shutdown implementation like
`net/http`.

Multiple third party packages try to provide shutdown capabilities by bringing
control over the server's `net.Listener` and relying on a shutdown event the
user handles in his application. This is not a viable approach: the user is not
given a way to drain connections and in-flight requests are eventually killed.

This package implements a mechanism similar to `net/http` to bring a proper
graceful shutdown feature by tracking connections and polling their state.

## Usage

```go
rpc.Register(new(MyInterface))
rpc.HandleHTTP()

l, err := net.Listen("tcp", ":1234")
if err != nil {
    log.Fatal("listen error:", err)
}

go func() {
    if err := http.Serve(l, nil); err != nil && err != rpc.ErrServerClosed {
        log.Fatal("serve error:", err)
    }
}()

...

// This will close all active listeners then wait for all RPC calls to finish
// before releasing connections and returning.
if err = rpc.Shutdown(context.Background()); err != nil {
    log.Fatal("shutdown error:", err)
}
```

## License

This project's inherits the golang project license.
