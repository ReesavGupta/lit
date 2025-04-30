# lit

**lit** is a simple and minimal implementation of Git, written in Go.

This project is my attempt to understand how Git works under the hood by building its core features from scratch.

## Goals

- Learn Git internals through hands-on implementation
- Build a basic version control system that mimics Git's core behavior

## Features (in progress)

- `init`: Initialize a new repository
- `cat-file`: Read Git objects (blobs)
- `hash-object`: Create Git blob objects
- `write-tree`, `read-tree`: Tree object operations
- `commit`: Create commits
- `clone`: Clone a remote repository
