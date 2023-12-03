package util

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

var cleanUpJobs = make(map[string]func() error)

// Should always be run in main function at the end. Other blocking processes in main function should be run in a go routine.
func SetCleanUp() {
	const cleanUpTimeOut = 10 * time.Second

	ch := make(chan os.Signal, 1)
	signal.Notify(ch, os.Interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	<-ch
	close(ch)

	wgCh := make(chan struct{})

	go func() {
		select {
		case <-time.After(cleanUpTimeOut):
			log.Fatalln("clean up timed out. Exited forcefully")
		case <-wgCh:
			close(wgCh)
			os.Exit(0)
		}
	}()

	wg := new(sync.WaitGroup)

	fmt.Println()
	for jobName, cleanUpJob := range cleanUpJobs {
		wg.Add(1)
		go func(jobName string, cleanUpJob func() error) {
			defer wg.Done()
			if err := cleanUpJob(); err != nil {
				fmt.Printf("✖  error running clean-up '%s': %s\n", jobName, err.Error())
			} else {
				fmt.Printf("✔  clean-up: %s\n", jobName)
			}
		}(jobName, cleanUpJob)
	}
	wg.Wait()
	wgCh <- struct{}{}
}

func RegisterCleanUp(jobName string, cleanUpJob func() error) {
	cleanUpJobs[jobName] = cleanUpJob
}
