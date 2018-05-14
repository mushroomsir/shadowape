# shadowape
[![Build Status](https://img.shields.io/travis/mushroomsir/shadowape.svg?style=flat-square)](https://travis-ci.org/mushroomsir/shadowape)
[![Coverage Status](http://img.shields.io/coveralls/mushroomsir/shadowape.svg?style=flat-square)](https://coveralls.io/github/mushroomsir/shadowape?branch=master)


# Usage

## server 
```json
{
    "server_config": {
        "server_addr": "0.0.0.0:8060",
        "cert_file": "",
        "key_file": ""
    }
}
```

go run main.go -c server.json

## client
```json
{
    "client_config": {
        "http_listen_addr": "0.0.0.0:8010",
        "socks5_listen_addr": "0.0.0.0:8020",
        "socks5_Server_addr": "127.0.0.1:8060",
        "root_cert_file": "testdata/root.pem"
    }
}
```

go run main.go -c client.json
