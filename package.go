package main

import (
	"fmt"
	"github.com/antchfx/htmlquery"
	"github.com/gookit/color"
)

// Build 构建
func (r *Tagger) Build(version Version) error {
	nonce, err := r.GetPackNonce()
	if err != nil {
		return err
	}

	color.Yellowln(fmt.Sprintf("开始构建版本 %s...", version.Name))

	resp, err := r.client.R().
		SetQueryParams(map[string]string{
			"page": "rosetta_manage_release_packages",
		}).
		SetFormData(map[string]string{
			"_wpnonce":         nonce,
			"_wp_http_referer": "/wp-admin/tools.php?page=rosetta_manage_release_packages",
			"action":           "build",
			"kind":             "glotpress",
			"source":           r.sourceByVersion(version.Name),
			"project":          r.projectByVersion(version.Name),
			"wp-rev":           version.Revision,
			"version":          version.Name,
			"submit":           "构建软件包",
		}).Post("https://cn.wordpress.org/wp-admin/tools.php")
	if err != nil {
		return err
	}

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return err
	}

	node := htmlquery.FindOne(doc, `//*[@id="message"]/p`)
	if node != nil {
		return fmt.Errorf(htmlquery.InnerText(node))
	}

	color.Greenln("构建成功！")

	return nil
}

// Release 发布
func (r *Tagger) Release(version string) error {
	color.Yellowln(fmt.Sprintf("开始发布版本 %s...", version))

	nonce, err := r.GetReleaseNonce(version)
	if err != nil {
		return err
	}

	resp, err := r.client.R().
		SetQueryParams(map[string]string{
			"page":     "rosetta_manage_release_packages",
			"action":   "release",
			"version":  version,
			"source":   r.sourceByVersion(version),
			"_wpnonce": nonce,
		}).Get("https://cn.wordpress.org/wp-admin/tools.php")
	if err != nil {
		return err
	}

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return err
	}

	node := htmlquery.FindOne(doc, `//*[@class="wp-die-message"]`)
	if node != nil {
		return fmt.Errorf(htmlquery.InnerText(node))
	}

	color.Greenln("发布成功！")
	return nil
}

// sourceByVersion 根据版本号获取源码
func (r *Tagger) sourceByVersion(version string) string {
	source := "trunk"
	// 6.5 版本开始 svn 上才有对应的分支
	if r.VersionCompare(version, "6.5", ">=") {
		source = "branches/" + version[:3]
	}
	return source
}

// projectByVersion 根据版本号获取源码
func (r *Tagger) projectByVersion(version string) string {
	source := "dev"
	if r.VersionCompare(version, MainVersion, "<") {
		source = fmt.Sprintf("%s.x", version[:3])
	}
	return "wp/" + source
}
