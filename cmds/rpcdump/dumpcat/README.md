# dumpcat

view RPC dump file.

## Install
```bash
go install github.com/fsgo/fsgo/cmds/rpcdump/dumpcat@master
```

## Useage
```bash
# dumpcat -help
Usage of dumpcat:
  -a string
    	filter action. r: Read, w:Write, c:Close; rc: Read and Close (default "rwc")
  -cid int
    	filter only which conn ID.
    	-1 : disable other conditions
    	0  : enable other  conditions
    	>0 : filter only this connID
    	 (default -1)
  -d	print details (default true)
  -s string
    	filter only which service
```