# postsocket

This repository contains a sketch of a Go implementation of the Transport
Services Abstract Interface described in
[draft-trammell-taps-interface](https://taps-api.github.io/drafts/draft-trammell-taps-interface.html).
This sketch is currently intended to illustrate issues and design choices made
during the development of that interface, and as a proof of concept that the
abstract interface is implementable in Go.

Future revisions of this repository will contain a full implementation of this
interface backed by TCP, UDP, and a userland QUIC implementation over UDP.
