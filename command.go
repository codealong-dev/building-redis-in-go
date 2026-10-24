package main

import (
	"codealong/store"
	"sync"
	"fmt"
)

// Command is a parsed Redis-ish operation: opcode plus its arguments.
type Command struct {
	Op    string
	Key   string
	Value string
}

// dispatch turns a Command into a store call.
func dispatch(s store.Storer, c Command) error {
	switch c.Op {
	case "SET":
		return s.Set(c.Key, c.Value)
	case "INCR":
		_, err := s.Incr(c.Key)
		return err
	default: 
		return fmt.Errorf("unknown op %q", c.Op)
	}
}

// RestoreOnBoot replays a slice of commands against the store, in parallel.
// Real Redis is single-threaded — we are not. This is where it gets fun.
func RestoreOnBoot(s store.Storer, cmds []Command) {
	var wg sync.WaitGroup
	errs := make(chan error, len(cmds))

	for _, c := range cmds {
		wg.Go(func() {
			errs <- dispatch(s, c)
		})
	}

	go func(){
		wg.Wait()
		close(errs)
	}()

	// log the error we get from a channel
	for err := range errs {
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		}
	}
}
