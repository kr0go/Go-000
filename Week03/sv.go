package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"

	"golang.org/x/sync/errgroup"
)

type t struct {
	port string
}

func (a t) ServeHTTP(resp http.ResponseWriter, req *http.Request) {
	fmt.Fprintln(resp, "handle", a.port)
}

func httpServ(ctx context.Context, port string, cancel context.CancelFunc) error {
	defer cancel()
	if r := recover(); r != nil {
		buf := make([]byte, 64<<10)
		buf = buf[:runtime.Stack(buf, false)]
		return fmt.Errorf(`errorgroup: panic recover:%s\n%s`, r, buf)
	}
	h := t{port: port}
	s := http.Server{
		Addr:    "127.0.0.1:" + port,
		Handler: http.Handler(h),
	}
	go func() {
		<-ctx.Done()
		fmt.Println("http with port " + port + " is ready to shutdown")
		s.Shutdown(ctx)
	}()
	return s.ListenAndServe()
}

// func badHttpServ(ctx context.Context, port string, cancel context.CancelFunc) error {
// defer cancel()
// if r := recover(); r != nil {
// buf := make([]byte, 64<<10)
// buf = buf[:runtime.Stack(buf, false)]
// return fmt.Errorf(`errorgroup: panic recover:%s\n%s`, r, buf)
// }
// h := t{port: port}
// s := http.Server{
// Addr:    "127.0.0.1:" + port,
// Handler: http.Handler(h),
// }
// go func() {
// time.Sleep(10 * time.Second)
// fmt.Println("http with port " + port + " is ready to shutdown")
// s.Shutdown(ctx)
// }()
// return s.ListenAndServe()
// }

func handleSignal(ctx context.Context, cancel context.CancelFunc, c chan os.Signal) error {
	select {
	case <-ctx.Done():
		fmt.Println("signal function got:", ctx.Err())
		return fmt.Errorf("signal function got: %v", ctx.Err())
	case <-c:
		fmt.Println("capture a signal:", <-c)
		cancel()
		return fmt.Errorf("capture a signal: %v", <-c)
	}
}

func main() {
	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGKILL, syscall.SIGQUIT, syscall.SIGTERM, syscall.SIGINT)
	ctx, cancel := context.WithCancel(context.Background())
	g, _ := errgroup.WithContext(ctx)
	g.Go(func() error {
		return httpServ(ctx, "8080", cancel)
	})
	g.Go(func() error {
		return httpServ(ctx, "8081", cancel)
	})
	// g.Go(func() error {
	// return badHttpServ(ctx, "8082", cancel)
	// })
	g.Go(func() error {
		return handleSignal(ctx, cancel, c)
	})
	if err := g.Wait(); err != nil {
		fmt.Println("In the end:", err)
	}
}
