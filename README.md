# utils
utils for go

### Build
**`need use -tags`**
```
go build -tags=[linux,windows,darwin] ...
```

### Tip!!!
If you get this when running your program:
```sh
fork/exec xxxxx: operation not permitted
```

Please use it as root or sudo.

or remove syscall.ProcAttr.Credential
or call SetNoSetGroups(true)
