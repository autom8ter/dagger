# dagger [![GoDoc](https://godoc.org/github.com/autom8ter/dagger?status.svg)](https://godoc.org/github.com/autom8ter/dagger)

Dagger is a collection of generic, concurrency safe datastructures.
Formerly solely a Graph library but now includes a number of additional datastructures.

    import "github.com/autom8ter/dagger"

## Supported Datastructures

All datastructures are concurrency safe by default. This means that all datastructures are safe to use in concurrent
applications without the need for external synchronization.
This is accomplished by using the `sync.RWMutex` in the background.
All datastructures are also generic. This means that all datastructures can be used with any type.

- [x] Directed Graph Datastructure
- [x] Stack(T) Datastructure
- [x] Queue(T) Datastructure
- [x] Set(T) Datastructure
- [ ] BTree(T) Datastructure (TODO)
- [ ] TTL Cache(T) Datastructure (TODO)
- [ ] LRU Cache(T) Datastructure (TODO)

Please see the GoDoc for more
information: [https://godoc.org/github.com/autom8ter/dagger](https://godoc.org/github.com/autom8ter/dagger)
