package daemon

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/hanifanggawi/vessel/internal/config"
)

const (
	TICK_INTERVAL_SECONDS = 60
)

func Run() error {
	eventChannel := make(chan Event, 4)
	reconcileChannel := make(chan struct{}, 1)

	ctx, cancel := context.WithCancel(context.Background())
	var waitgroup sync.WaitGroup

	waitgroup.Add(1)
	go signalHandler(ctx, &waitgroup, eventChannel)
	go reconciler(ctx, &waitgroup, reconcileChannel)
	go scheduler(ctx, &waitgroup, eventChannel)

	coordinator(ctx, cancel, eventChannel, reconcileChannel)

	waitgroup.Wait()
	fmt.Println("vessel daemon exited")
	return nil
}

func signalHandler(ctx context.Context, wg *sync.WaitGroup, eventCh chan<- Event) {
	defer wg.Done()

	sigsCh := make(chan os.Signal, 1)
	signal.Notify(sigsCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)
	defer signal.Stop(sigsCh)

	for {
		select {
		case <-ctx.Done():
			return
		case sig := <-sigsCh:
			switch sig {
			case syscall.SIGHUP:
				eventCh <- Event{Kind: EventReload}
			case syscall.SIGTERM, syscall.SIGINT:
				eventCh <- Event{Kind: EventShutdown}
			}
		}
	}
}

func scheduler(ctx context.Context, wg *sync.WaitGroup, eventCh chan<- Event) {
	defer wg.Done()

	ticker := time.NewTicker(time.Duration(TICK_INTERVAL_SECONDS) * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			eventCh <- Event{Kind: EventTick}
		}
	}
}

func reconciler(ctx context.Context, wg *sync.WaitGroup, reconcileChannel <-chan struct{}) {
	defer wg.Done()

	for {
		select {
		case <-ctx.Done():
			return
		case <-reconcileChannel:
			if err := config.RunReconcile(); err != nil {
				fmt.Println(err.Error())
			} else {
				fmt.Println("Reconcile Success")
			}
		}
	}
}

func coordinator(ctx context.Context, cancel context.CancelFunc, eventCh <-chan Event, reconcileCh chan<- struct{}) {
	reconcileCh <- struct{}{}

	for {
		select {
		case <-ctx.Done():
			return
		case event := <-eventCh:
			switch event.Kind {
			case EventTick:
				select {
				case reconcileCh <- struct{}{}:
				default:
					fmt.Println("Reconciler busy, skipping")
				}
			case EventReload:
				reconcileCh <- struct{}{}
			case EventShutdown:
				cancel()
				return
			}
		}
	}
}
