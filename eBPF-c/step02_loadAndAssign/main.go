package main

import (
	"log"
	"os"
	"os/signal"
	"time"

	"github.com/cilium/ebpf"
)

func main() {
	spec, err := ebpf.LoadCollectionSpec("s2.bpf.o")
	if err != nil {
		panic(err)
	}

	// Instantiate a Collection from a CollectionSpec.
	coll, err := ebpf.NewCollection(spec)
	if err != nil {
		panic(err)
	}
	// Close the Collection before the enclosing function returns.
	defer coll.Close()

	// Obtain a reference to 'my_map'.
	m := coll.Maps["my_map"]

	// Set map key '1' to value '2'.
	if err := m.Put(uint32(1), uint64(2)); err != nil {
		panic(err)
	}
    var result uint64
    m.Lookup(uint32(1), &result)
    log.Printf("[%+v]",result)


    var count1,count2 uint64    
    tick := time.Tick(time.Millisecond)
	stop := make(chan os.Signal, 5)
	signal.Notify(stop, os.Interrupt)
	for {
		select {
		case <-tick:			
			err := m.Put(uint32(2), &count1)            
			if err != nil {
				log.Fatal("Map lookup:", err)
			}
            count1++
            err=m.Lookup(uint32(2), &count2)
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
