[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)

# gobackgammon

golang is a perfect language for a program that does Monte Carlo simulations of
backgammon positions to determine the expected value of points gained for a
player. That's because (1) it can easily utilize all available CPU hardware for
the computation and (2) the generated code is not as fast as C/C++ but
significantly faster than Python/Ruby.

## Copyright

Copyright 2018 David L. Chandler

See the LICENSE file in this directory.

## FAQ

Q: Is it any good?
A: Oh yes.

## What's next

* `git grep TODO` will show you some future directions, especially the use of
Monte-Carlo simulations to improve the AI.

* TODO(chandler37): Implement a method that takes a brd.Board and generates a
scalable vector graphics (SVG) representation of it.

* TODO(chandler37): Building on that SVG, make a web server that you can use
play a game against the AI.

* I'd like to use https://github.com/jamiealquiza/tachymeter to measure how long
it takes to have ai.MakePlayerConservative(0, nil) play a game against
itself. We have `make bench` right now to give us an idea of performance, but
it seems like the variance is surprisingly high.
