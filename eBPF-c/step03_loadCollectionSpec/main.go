package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/cilium/ebpf"
)

type myObjs struct {
	MyMap  *ebpf.Map     `ebpf:"my_map"`
	MyProg *ebpf.Program `ebpf:"my_prog"`
}

func (objs *myObjs) Close() error {
	if err := objs.MyMap.Close(); err != nil {
		return err
	}
	if err := objs.MyProg.Close(); err != nil {
		return err
	}
	return nil
}

func main() {
	spec, err := ebpf.LoadCollectionSpec("s3.bpf.o")
	if err != nil {
		panic(err)
	}
	// Look up the __64 type declared in linux/bpf.h.
	t, err := spec.Types.AnyTypeByName("__u64")
	if err != nil {
		panic(err)
	}
	fmt.Println(t)

	var objs myObjs
	if err := spec.LoadAndAssign(&objs, nil); err != nil {
		panic(err)
	}
	defer objs.Close()

	// Interact with MyMap through the custom struct.
	if err := objs.MyMap.Put(uint32(1), uint64(2)); err != nil {
		panic(err)
	}

	var result uint64
	objs.MyMap.Lookup(uint32(1), &result)
	log.Printf("[%+v]", result)

	var count1, count2 uint64
	tick := time.Tick(time.Millisecond)
	stop := make(chan os.Signal, 5)
	signal.Notify(stop, os.Interrupt)
	for {
		select {
		case <-tick:
			err := objs.MyMap.Put(uint32(2), &count1)
			if err != nil {
				log.Fatal("Map lookup:", err)
			}
			count1++
			err = objs.MyMap.Lookup(uint32(2), &count2)
			if err != nil {
				log.Fatal("Map lookup:", err)
			}
			log.Printf("[%+v]", count2)

		case <-stop:
			log.Print("Received signal, exiting..")
			return
		}
	}
}
