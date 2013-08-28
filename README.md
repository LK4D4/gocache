gocache
=======

Simple caching application, like memcached

Install
-------

Clone repository and change dir:
```
git clone https://github.com/LK4D4/gocache && cd gocache
```

Run tests:
```
make test
```

Install package:
```
make
```

Try
---

Run gocache:
```
bin/gocache -port 6090 -ncpu 4
```

Try it:
```
telnet 127.0.0.1 6090
Trying 127.0.0.1...
Connected to 127.0.0.1.
Escape character is '^]'.
set a 5
OK
get a
OK 5
```

Run benchmark:
```
bin/bench -v 4 -host 127.0.0.1 -port 6090
```
