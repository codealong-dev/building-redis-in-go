package main

import (
	"codealong/store/kv"
	"fmt"
)

func main() {
	s := kv.NewStore(0)

	s.Set("pageviews", "0")

	const hits = 1000
	cmds := make([]Command, hits)
	for i := range cmds {
		cmds[i] = Command{Op: "INCR", Key: "pageviews"}
	}

	cmds = append(cmds,
		Command{Op: "INCR", Key: ""},          // empty key -> ErrEmptyKey
		Command{Op: "WAT", Key: "pageviews"},  // unknown op
	)

	RestoreOnBoot(s, cmds)

	got, _ := s.Get("pageviews")
	fmt.Printf("expected: %d\n", hits)
	fmt.Printf("got:      %s\n", got)
}
