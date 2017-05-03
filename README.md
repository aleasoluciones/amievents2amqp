# amievents2amqp

## Deployment
This component is deployed within the issues component. You need to build and release a new image:

```
$ ssh -t asur@mamut release.sh -r amievents2amqp -g `git rev-parse --short HEAD`
```

Then go to the issues repository, update the amievents2amqp image tag to the last commit in docker-compose, and deploy issues.

## Development

To connect to the asterisk manager you need to grant access from your ip.

First connect to the asterisk machine via ssh:

```
$ ssh asur@ale-niv1
```

Edit the manager configuration:

```
$ sudo vim /opt/bos/asterisk/etc/manager_custom.conf
```

It looks similar to:

```
[felix]
secret = <the-password>
deny = 0.0.0.0/0.0.0.0
permit = 192.168.10.14/255.255.255.255,192.168.10.35/255.255.255.255
read = system,call,log,verbose,command,agent,user
write = system,call,log,verbose,command,agent,user
```

Just add a new item to the `permit` line and restart asterisk:

```
$ sudo asterisk -r
asterisk> core restart when convenient
```

*WARNING:* Do not forget to remove your ip after development!
