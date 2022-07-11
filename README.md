[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# gobackgammon

golang is a perfect language for a program that does Monte Carlo simulations of
backgammon positions to determine the expected value of points gained for a
player. That's because (1) it can easily utilize all available CPU hardware for
the computation and (2) the generated code is not as fast as C/C++ but
significantly faster than Python/Ruby.

See also the webapp built on top of this, https://github.com/chandler37/gobackgammond

## Copyright

Copyright 2018 David L. Chandler

See the LICENSE file in this directory.

## FAQ

Q: Is it any good?
A: Oh yes.

## What's next

* `git grep TODO` will show you some future directions, especially the use of
Monte-Carlo simulations to improve the AI.

* I'd like to use https://github.com/jamiealquiza/tachymeter to measure how long
it takes to have ai.MakePlayerConservative(0, nil) play a game against
itself. We have `make bench` right now to give us an idea of performance, but
it seems like the variance is surprisingly high.


## What do I type?

Don't type anything; instead, start with
https://github.com/chandler37/gobackgammond which has a webserver to help
visualize the code in this module.

But if you really want to modify this module:
- read ./Makefile
- install Go
  - MacOS? Run `brew install go` (see https://brew.sh/) to install golang. If
    your golang is out of date, `brew upgrade go`.
- run 'make check bench`
- publish a new git tag for a new version, say v1.m.p
- change https://github.com/chandler37/gobackgammond go.mod to reference v1.m.p and run `make clean srv`
