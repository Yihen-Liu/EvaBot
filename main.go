/**
 * Description:
 * Author: Yihen.Liu
 * Create: 2021-04-30
 */
package main

import (
	"github.com/riversgo007/EvaBot/features/captcha"
	"github.com/riversgo007/EvaBot/features/invite"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main()  {
	go invite.RunService()
	go captcha.RunService()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan
	log.Println("Shutdown signal received, exiting...")
}