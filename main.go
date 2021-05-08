/**
 * Description:
 * Author: Yihen.Liu
 * Create: 2021-04-30
 */
package main

import (
	"github.com/riversgo007/EvaBot/features/captcha"
	"github.com/riversgo007/EvaBot/features/invite"
)

func main()  {
	go captcha.RunService()
	go invite.RunService()
	select {}
}