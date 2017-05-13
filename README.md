# ping

Your very own Google Analytics replacement, without all of the Google.
Simple as pie.

## Motivation

Google Analytics is creepy. It knows where you are, what computer you're
using, what browser you're using, what page you visited, and so on. It has
horizontal data for IP addresses, so Google knows what sites you've visited
across the Internet, for how long, and what your path was. I'm
uncomfortable giving Google all this data. In fact, I even wrote [a blog
post about how to block the various tracking services.](http://blog.parkermoore.de/2014/07/16/dont-like-being-tracked/)

**But**, I wanted to know what my "greatest" hits are. I wanted to see what
people liked and didn't like. Using Google Anayltics isn't an option, but
what about using something much simpler? Enter: ping.

## Ping: What It Is

Ping is a tiny little server that logs three things:

- IP address of visitor
- URL they visited
- When the visit happened

A single tiny JavaScript file that the user requests sends down these three
things and that's all there is to it. As unintrusive as possible, while
still providing insight into the site's strenghts.

## Installation

Want to run ping? No problem.

```bash
$ go get github.com/parkr/ping
$ mysql -e 'create database ping;'
$ PING_DB=user:passwd@localhost/ping ping -http=:8972
```

Specify a port (defaults to `8000`) and a database URL and you're off to
the races. Enjoy!

Running behind a proxy? No problem. Specify `PING_PROXIED=true` when
invoking `ping` and you're good to go.

Prefer Docker? We got that too!

```bash
$ mysql -e 'create database ping;'
$ docker run --rm \
  -e PING_DB=user:passwd@host-ip/ping \
  parkr/ping \
  ping -http=:8972
```

Ensure that your Docker container is given access to the server MySQL is
running on and that MySQL's `bind-address` host is `0.0.0.0` so it can be
accessed by others. Or, use a MySQL proxy.
