# Nettee

Nettee implements a network tee used to multiplex telnet/terminal server streams among multiple clients.

~~~~
$ nettee 
  -localAddress string
    	local address
  -remoteAddress string
    	remote address
  -timeout duration
    	remote connection timeout (default 1h0m0s)

$ nettee -localAddress :8080 -remoteAddress your_ts:4039 -timeout 50m
~~~~
