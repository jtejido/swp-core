# Transport Binding T1: HTTP/2 Streams (Informative)

## 1. Scope

This document describes an optional transport binding for carrying SWP frames over HTTP/2 streams.

## 2. Overview

- one HTTP/2 stream carries a sequence of SWP frames
- frames are carried in HTTP/2 DATA frames
- recommended content-type: `application/swp`

## 3. Identity and security

When paired with S1 security binding:

- peer identity is derived from the secure channel
- identity is surfaced to profile handlers for authorization

## 4. Operational notes

- HTTP/2 flow control provides backpressure
- SWP defines no additional flow-control mechanism at this layer

