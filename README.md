## bytebuf

UPDATE: This implementation is deprecated as the patch is merged to the Go mainline (https://golang.org/cl/133715)

Example of how [CL133375 cmd/compile/internal/gc: handle array slice self-assign in esc.go](https://golang.org/cl/133375)
can be utilized to mitigate Go escape analysis limitations.

It shows how to fix problem described in
[cmd/compile, bytes: bootstrap array causes bytes.Buffer to always be heap-allocated](https://github.com/golang/go/issues/7921)
using brand new escape analysis pattern.

## Implementation difference

The whole implementation difference can be described as:

```diff
type Buffer struct {
	buf       []byte   // contents are the bytes buf[off : len(buf)]
	off       int      // read at &buf[off], write at &buf[len(buf)]
- 	bootstrap [64]byte // memory to hold first slice; helps small buffers avoid allocation.
+ 	bootstrap *[64]byte // memory to hold first slice; helps small buffers avoid allocation.
	lastRead  readOp   // last read operation, so that Unread* can work correctly.
}
```

With updated escape analysis, it's possible to actually take benefits of
bootstrap array, but only if it's not "inlined" into `Buffer` object.
So, we need a pointer to array instead of normal array.

This makes it impossible to use zero value though, hence `New` function.

## Performance comparison before [bytes.Buffer optimization in CL133715](https://golang.org/cl/133715)

> These results provided only for historical reference.

Given this code:

```
buf.Write(s)
globalString = buf.String()
```

Where `s` is:

| Label | Data |
|-------|------|
| empty | empty string  |
| 5     | 5-byte string |
| 64    | 64-byte string |
| 128   | 128-byte string |
| 1024  | 1024-byte string |

```
name            old time/op    new time/op    delta
String/empty-8     138ns ±13%      24ns ± 0%   -82.94%  (p=0.000 n=10+8)
String/5-8         186ns ±11%      60ns ± 1%   -67.82%  (p=0.000 n=10+10)
String/64-8        225ns ±10%     108ns ± 6%   -52.26%  (p=0.000 n=10+10)
String/1024-8      889ns ± 0%     740ns ± 1%   -16.78%  (p=0.000 n=9+10)

name            old alloc/op   new alloc/op   delta
String/empty-8      112B ± 0%        0B       -100.00%  (p=0.000 n=10+10)
String/5-8          117B ± 0%        5B ± 0%   -95.73%  (p=0.000 n=10+10)
String/64-8         176B ± 0%       64B ± 0%   -63.64%  (p=0.000 n=10+10)
String/1024-8     2.16kB ± 0%    2.05kB ± 0%    -5.19%  (p=0.000 n=10+10)

name            old allocs/op  new allocs/op  delta
String/empty-8      1.00 ± 0%      0.00       -100.00%  (p=0.000 n=10+10)
String/5-8          2.00 ± 0%      1.00 ± 0%   -50.00%  (p=0.000 n=10+10)
String/64-8         2.00 ± 0%      1.00 ± 0%   -50.00%  (p=0.000 n=10+10)
String/1024-8       3.00 ± 0%      2.00 ± 0%   -33.33%  (p=0.000 n=10+10)
```

## Performance comparison after [bytes.Buffer optimization in CL133715](https://golang.org/cl/133715)

After improvements to `bytes.Buffer` from the standard library and
updated tests that do `bytes.NewBuffer(make([]byte, 0, 64))`,
results are mostly in favor of `bytes.Buffer`.

Also note that `bytes.Buffer` makes it possible to have a custom size
stack-allocated slice as a bootstrap storage, so it's overall a
better choice.

```
name            old time/op    new time/op    delta
String/empty-8    23.8ns ± 0%    25.7ns ± 0%   +8.04%  (p=0.000 n=8+8)
String/5-8        59.6ns ± 1%    53.4ns ± 1%  -10.32%  (p=0.000 n=10+10)
String/64-8        102ns ±12%      95ns ± 4%   -6.94%  (p=0.001 n=10+8)
String/1024-8      745ns ± 0%     768ns ± 1%   +3.08%  (p=0.000 n=9+8)

name            old alloc/op   new alloc/op   delta
String/empty-8     0.00B          0.00B          ~     (all equal)
String/5-8         5.00B ± 0%     5.00B ± 0%     ~     (all equal)
String/64-8        64.0B ± 0%     64.0B ± 0%     ~     (all equal)
String/1024-8     2.05kB ± 0%    2.18kB ± 0%   +6.25%  (p=0.000 n=10+10)

name            old allocs/op  new allocs/op  delta
String/empty-8      0.00           0.00          ~     (all equal)
String/5-8          1.00 ± 0%      1.00 ± 0%     ~     (all equal)
String/64-8         1.00 ± 0%      1.00 ± 0%     ~     (all equal)
String/1024-8       2.00 ± 0%      2.00 ± 0%     ~     (all equal)
```
