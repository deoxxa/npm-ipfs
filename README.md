npm-ipfs
========

This is an experiment in glueing npm to IPFS via a small web server, while
still providing the same semver functionality as the official npm registry.
The reason this is implemented as a separate server is that I didn't feel like
digging too far into npm itself just to see if this is possible.

This same functionality could be provided as a part of the npm tool itself,
or as a part of an alternate implementation. I'd suggest looking at [ied](https://github.com/alexanderGugel/ied)
if you want to go down that route.

Problems Addressed
------------------

The problems that I have with npm are as follows:

1. conflict of interest - the npm registry is entirely owned and controlled by
   one entity whose success relies, in part, on maintaining control of where
   packages come from in general. This means there's very little incentive for
   npm inc to provide any federation opportunities for running private or
   independent infrastructure in a cooperative way with the main registry.
2. single point of failure (usually) - most people use the main registry.
   There are very few replicas outside of large companies or those with good
   ops people. Smaller organisations with fewer people are hit hardest when
   something goes wrong with the main registry (downtime, drama, etc).
3. package deletion/withdrawl - the npm registry is the worst kind of mutable
   structure. It doesn't completely protect consumers by disallowing deletes,
   but it also doesn't serve those who'd like to delete things by making it
   easy. It's too easy to think of the npm registry as immutable, where it
   really isn't.

The first point is obviously moot if packages are primarily controlled by
their authors, with fallback to the community. Thanks, IPFS!

The second and third points can be addressed by mechanisms in IPFS such as
pinning, where certain files (e.g. packages you have used) are kept in a local
cache. Once enough people are using a package, there will be copies of it in
many places, and it won't disappear. At the same time, the author can disclaim
ownership of a package by ceasing to provide metadata for it via their IPNS
name.

There are still some holes to be poked in this idea, but I think it's got
potential.

Quickstart
----------

The address below is significant. In this early stage, I've hard-coded the
host for the dependencies in package.json to point to `127.0.0.1:3001`.

```
$ go build
$ ipfs daemon &
$ ./npm-ipfs --addr :3001
```

(then in another terminal)

```
$ mkdir node_modules
$ npm install http://127.0.0.1:3001/QmNYdjhpNii2zh2B3iJTywvqRg1Ub1U7Ww8sDony7z9v8z/testing@1.0.0
$ npm ls
```

Check out ipns/QmNYdjhpNii2zh2B3iJTywvqRg1Ub1U7Ww8sDony7z9v8z to see how I
have packages set up. The directory structure is mirrored in the `packages`
directory of this repository.

You can create packages like that with npm itself - run `npm pack` and it'll
give you a packed tarball. Rename it so that it looks like `name@a.b.c` instead
of `name-a.b.c`, publish it on IPFS, and you can install it straight away.
