# BugList

---
1. gobler - sender goroutine panics when receiving nil message.MessagePack gob

Reproduction:

* run `make`, new spool_test.go code creates an empty mp `&message.MessagePack{}` and drops its gob into /tmp
* doesn't delete it after test, making endly test_02 run into it and panics
  * spool_test now does MkdirTemp() and cleans up afterwards, so this bug doesn't happen
  * also, there are checks in place before sender routine that prevent this from happening in real-world
  * still, this should not happen, figure out how to protect gobler from such malformed "injections"

```
[1]+  Exit 2                  nohup gobler -c /tmp/gobler.conf
[run[loop_over_tests01]run|[run_gobler]process.start nohoup                                                                                                                                                stdout]
2022/08/12 22:26:37 Initializing connector: discord
2022/08/12 22:26:37 Initializing connector: mailto
2022/08/12 22:26:37 Initializing connector: matrix
2022/08/12 22:26:37 Initializing connector: msteams
2022/08/12 22:26:37 Initializing connector: telegram
panic: runtime error: invalid memory address or nil pointer dereference
[signal SIGSEGV: segmentation violation code=0x1 addr=0x318 pc=0x9468cd]

goroutine 10 [running]:
main.(*sender).SenderWorker(0xc0002aa940, 0x0?, 0x0?, 0x0?, 0x0?)
        /home/pja/src/go-projects/goslmailer-clip-hpc/cmd/gobler/sender.go:58 +0x3cd
created by main.(*conMon).SpinUp
        /home/pja/src/go-projects/goslmailer-clip-hpc/cmd/gobler/conmon.go:151 +0x55b
```

---
