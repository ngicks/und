package main

// intersection
//   - Option:
//     - def&&(null||und) => no conversion
//     - def => T (Value method)
//     - null||und => *struct{}
//   - Und:
//     - def&&null&&und => no conversion
//     - def&&null => Option[T]
//     - def&&und => Option[T]
//     - null&&und => Option[*struct]
//     - def => T (Value method)
//     - null||und => *struct{}
//   - Elastic:
//     - def => T (Value method)
//     - def&&null => Option[T]
//     - def&&und => Option[T]
//     - null&&und => Option[*struct]
//     - null||und => *struct{}
//     - def&&null&&und => no conversion
//     - If len==n, then T is [n]U
//     - If len<n,len<=n,len>n,len>=n, then T is []U where len(u) is at most / at least n.
//     - If values:nonull is not set, U is option.Option[R]
