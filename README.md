# amievents2amqp

This program connects to an remote Asterisk server by TCP, then it asks Asterisk to receive the Asterisk generated event. Then for each event, it publish it to the given AMQP broker.

```
Usage of /amievents2amqp:
  -amiURI string
        AMI connection URI (use AMI_URI env var as default)
  -amqpURI string
        AMQP connection URI (use BROKER_URI env var as default)
  -exchange string
        AMQP exchange name (default "events")
  -topic string
        topic (default "astevents")

```

## Deployment

This component is deployed within the issues component. You need to build and release a new image:

```
$ ssh -t asur@mamut release.sh -r amievents2amqp -g `git rev-parse --short=7 HEAD`
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
