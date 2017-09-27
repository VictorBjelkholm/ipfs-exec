## ipfs-exec
> Executes commands in new repository and version

### Usage

Automatically downloads and inits if missing version.

```
ipfs-exec $VERSION $COMMAND

ipfs-exec v0.4.11-rc2 daemon
ipfs-exec v0.4.11-rc2 add file.md

ipfs-exec v0.4.11-rc1 daemon
ipfs-exec v0.4.11-rc1 cat $HASH
```

Now you'd have two repository in `~/.ipfs-exec` directory, with two running daemons
on different ports.

### Todo

- [ ] Massage and clean up some code
- [ ] Tests eh
- [ ] Set custom dist url
- [ ] Make work for more platforms than OSX (Windows, Linux)
