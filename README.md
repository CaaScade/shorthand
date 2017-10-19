# Shorthand
-------------

Encode your domain-specific knowledge and automatically generate a user-friendly API.

Applied to Kubernetes, it allows us to provide an ergonomic spec format (and tools!)
that's simpler to learn and easier to get right.

### Make Kubernetes Spec more readable

### Completely reversible

### Getting Started

```
Sub Commands
-------------

shrink - simplifies Kubernetes Spec file
grow - reverses the simplification
```
In order to build it, just run `./scripts/build`

Then, you can use it

```
./shorthand shrink SOURCE_DIR|FILE [OUTPUT_DIR] 

Note that if OUTPUT_DIR is empty, then it outputs the file to current directory. 
Shorthand will never overwrite any existing files.
```
