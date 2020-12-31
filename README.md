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

It also respects the [Do Not Track header](http://donottrack.us/), which
many browsers now allow users to set for all requests. Complying with this
header is not mandatory, but aligns nicely with our motivation for
respecting users' privacy when they ask for it.

## Installation

Want to run ping? No problem.

```bash
$ go get github.com/parkr/ping
$ PING_DB=./ping_production.sqlite3 ping -http=:8972
```

Specify a port (defaults to `8000`) and a database URL and you're off to
the races. Enjoy!

Running behind a proxy? No problem. Specify `PING_PROXIED=true` when
invoking `ping` and you're good to go.

Prefer Docker? We got that too!

```bash
$ mkdir data
$ docker run --rm \
  -e PING_DB=/srv/data/ping_production.sqlite3 \
  -v $(pwd)/data:/srv/data:rw \
  parkr/ping \
  ping -http=:8972
```

This will save all data to the specified sqlite3 database,
mounted to the container and written back to the host.

Then, load the script on your pages in the HTML:

```html
<script src="https://domain.for.ping.server/ping.js" type="text/javascript"></script>
```

Every page will load this `/ping.js` path from the `ping` server and will
log the result to your sqlite3 database in the `visits` table.

If you set the `-hosts` flag for the `ping` executable, you can limit which
hostnames will be logged. This is a security feature to ensure your sqlite3
database doesn't bloat with unwanted data from sites you don't want to
track. For example, adding the flag `-hosts=example.com,my-site.com,mysite.rocks`
will only log visits to your database for visits to `example.com`,
`my-site.com`, and `mysite.rocks`. Of course, these sites have to load the
javascript path as specified above, so this will only work for sites you
control.
