# dumpreplay


Replay RPC Dump Data to server.

## Install
```bash
go install github.com/fsgo/fsgo/cmds/rpcdump/dumpreplay@master
```

## Useage

```bash
# dumpreplay -help
Usage of dumpreplay:
  -cid int
    	filter only which conn ID
  -conc int
    	Number of multiple requests to make at a time (default 1)
  -dist string
    	replay data to
  -s string
    	filter only which service
```