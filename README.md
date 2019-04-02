# prometheus-fping

This is a very simple wrapper around [fping][fping] as a prometheus exporter.

The only configuration option available is `NETWORK`, which defines the network
for fping to scan, you can configure that in
`/etc/prometheus-fping/environment`

    $ cat /etc/prometheus-fping/environment
    NETWORK=192.168.1.0/24
