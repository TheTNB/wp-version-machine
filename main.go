package main

import (
	"os"
	"time"

	"github.com/gookit/color"
	"github.com/imroc/req/v3"
	_ "github.com/joho/godotenv/autoload"
)

const MainVersion = "6.5"

type Tagger struct {
	username        string
	password        string
	yescaptchaToken string
	client          *req.Client
}

func main() {
	username := os.Getenv("WPORG_USERNAME")
	password := os.Getenv("WPORG_PASSWORD")
	yescaptchaToken := os.Getenv("YESCAPTCHA_TOKEN")
	if username == "" || password == "" {
		color.Redln("请先在 .env 文件中设置 WPORG_USERNAME 和 WPORG_PASSWORD 为 WordPress.org 的用户名和密码")
		return
	}
	if yescaptchaToken == "" {
		color.Redln("请先在 .env 文件中设置 YESCAPTCHA_TOKEN 为 YesCaptcha 的 Token 用于识别验证码")
		return
	}

	client := req.C()
	client.SetTimeout(3 * time.Minute)
	client.SetCommonRetryCount(2)
	client.ImpersonateChrome()
	client.EnableInsecureSkipVerify()

	tagger := &Tagger{
		username:        username,
		password:        password,
		yescaptchaToken: yescaptchaToken,
		client:          client,
	}

	if err := tagger.Login(); err != nil {
		color.Redln(err.Error())
		return
	}

	go tagger.KeepLogin()
	go tagger.FetchUpdate()

	color.Greenln("定时器已启动！")

	select {}
}
