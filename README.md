# XferTool 

## HTTPS To Multiple HTTP Backends ##

### Why ###
I was trying to setup a couple different sub-domains (using Go programs), and wanted a different Go program to deal with each sub-domain. Since I was using a 
_Let's Encrypt_ wildcard cert, it wouldn't let me run the cert on any ports other than 443.  There is apparently some kind of security concern with running
the same certificate on multiple ports, and since you can't bind more than one program to a port, this was becoming a real pain.  I tried to front the cert with
nginx, and proxy_pass to the backend Go programs, but any _POST_ data was somehow lost in the proxy_pass transfer.  That lead to writing the XferTool, which 
fronts the HTTPS connection, and (based on "Host:" string of the HTTPS header) redirect to whatever port you have specified in the config.

### How It's Done ###
Basically, XferTool deals with all of the TLS work up front, reads the initial request header, loops through the list provided in the config file, and tries
to match what the browser is requesting with a domain / port mapping in the config file.  There is a default port option in the config so that any unmatched
domain connections get routed there.

### Example Config ###
```
# Path To Your Certs
# TLS In Go Requires The FullChain, Not The cert.pem Or ca.crt
FullChainCertPath: "/etc/letsencrypt/live/somedomain/fullchain.pem"
PrivKeyCertPath: "/etc/letsencrypt/live/somedomain/privkey.pem"

# The Default Port To Proxy To If Domain Isn't Listed
DefaultPort: 9090

# Domain To Port Mappings
Domains:
  dev.yoursite.com: '9091'
  yoursite.com: '9090'
  www.yoursite.com: '9090'
```
