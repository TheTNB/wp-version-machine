package main

import (
	"fmt"
	"github.com/gookit/color"
	"regexp"
	"strings"

	"github.com/antchfx/htmlquery"
)

func (r *Tagger) GetPackNonce() (string, error) {
	color.Greenln("获取构建软件包的 nonce...")
	resp, err := r.client.R().
		Get("https://cn.wordpress.org/wp-admin/tools.php?page=rosetta_manage_release_packages")
	if err != nil {
		return "", err
	}

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return "", err
	}

	node := htmlquery.FindOne(doc, `//*[@id="_wpnonce"]`)
	if node == nil {
		return "", fmt.Errorf("未找到 nonce")
	}

	nonce := htmlquery.SelectAttr(node, "value")
	color.Greenln("取得 nonce: " + nonce)
	return nonce, nil
}

// GetReleaseNonce 获取发布版本的 nonce，这个每个版本号都不一样，实测不能复用
func (r *Tagger) GetReleaseNonce(version string) (string, error) {
	color.Greenln(fmt.Sprintf("获取版本 %s 的 nonce...", version))
	resp, err := r.client.R().
		Get("https://cn.wordpress.org/wp-admin/tools.php?page=rosetta_manage_release_packages")
	if err != nil {
		return "", err
	}

	reg := regexp.MustCompile(`<a\sclass="edit"\shref="/wp-admin/tools.php\?page=rosetta_manage_release_packages&#038;action=release&#038;version=([^&]+)&#038;source=([^&]+)&#038;_wpnonce=([^&]+)">`)
	matches := reg.FindAllStringSubmatch(resp.String(), -1)

	for _, match := range matches {
		if strings.TrimSpace(match[1]) == version {
			color.Greenln("取得 nonce: " + match[3])
			return match[3], nil
		}
	}

	return "", fmt.Errorf("未找到版本 %s 的 nonce", version)
}
