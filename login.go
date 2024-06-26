package main

import (
	"github.com/antchfx/htmlquery"
	"github.com/gookit/color"
)

func (r *Tagger) Login() error {
	color.Greenln("用户名: " + r.username)
	color.Greenln("尝试登录 WordPress.org...")
	captcha, err := r.GetRecaptcha()
	if err != nil {
		return err
	}

	resp, err := r.client.R().
		SetFormData(map[string]string{
			"log":                 r.username,
			"pwd":                 r.password,
			"rememberme":          "forever",
			"wp-submit":           "Log In",
			"redirect_to":         "https://cn.wordpress.org/wp-admin/",
			"_reCaptcha_v3_token": captcha,
		}).Post("https://login.wordpress.org/wp-login.php")
	if err != nil {
		return err
	}
	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return err
	}
	errnode := htmlquery.FindOne(doc, `//*[@id="login_error"]/p`)
	if errnode != nil {
		color.Redf("%s\n", htmlquery.InnerText(errnode))
	}

	return nil
}
