Proof of concept fast GeoIP2 microservice. Output format and code is likely to
change significantly. Use at your own risk.

Currently this requires the `greg/network-offset` branch on `oschwald
/maxminddb-golang`.

This is _not_ optimized for overall memory usage, but it is optimized to
minimize the number of mallocs.
