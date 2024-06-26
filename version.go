package main

import (
	"fmt"
	"strings"

	"github.com/antchfx/htmlquery"
	"github.com/spf13/cast"
)

type Version struct {
	Name     string `json:"name"`
	Revision string `json:"revision"`
}

// Versions 解析 WordPress 的 SVN 版本列表
func (r *Tagger) Versions() ([]Version, error) {
	resp, err := r.client.R().
		SetQueryParams(map[string]string{
			"order": "date",
			"desc":  "1",
		}).
		Get("https://build.trac.wordpress.org/browser/tags")
	if err != nil {
		return nil, err
	}

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	node := htmlquery.Find(doc, `//*[@id="dirlist"]/tbody/tr`)
	if node == nil {
		return nil, fmt.Errorf("未找到版本列表")
	}

	var versions []Version
	for i, n := range node {
		if i == 0 {
			continue
		}
		name := htmlquery.FindOne(n, "td[1]/a")
		ver := htmlquery.FindOne(n, "td[3]/a[1]")
		if name == nil || ver == nil {
			continue
		}
		versions = append(versions, Version{
			Name:     strings.TrimSpace(htmlquery.InnerText(name)),
			Revision: strings.TrimSpace(htmlquery.InnerText(ver)),
		})
	}

	return versions, nil
}

// PublishedVersions 解析 WordPress 后台已经发布的版本列表
func (r *Tagger) PublishedVersions() ([]Version, error) {
	resp, err := r.client.R().
		Get("https://cn.wordpress.org/wp-admin/tools.php?page=rosetta_manage_release_packages")
	if err != nil {
		return nil, err
	}

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	node := htmlquery.Find(doc, `/html/body/div[1]/div[2]/div[2]/div[1]/div[2]/table[2]/tbody/tr`)
	if node == nil {
		return nil, fmt.Errorf("未找到版本列表")
	}

	var versions []Version
	for _, n := range node {
		name := htmlquery.FindOne(n, "td")
		if name == nil {
			continue
		}
		// 跳过含有 show-more class 的
		if strings.Contains(htmlquery.SelectAttr(name, "class"), "show-more") {
			continue
		}
		versions = append(versions, Version{
			Name: strings.TrimSpace(htmlquery.InnerText(name)),
		})
	}

	return versions, nil
}

// BuiltVersions 解析 WordPress 后台已经构建的版本列表
func (r *Tagger) BuiltVersions() ([]Version, error) {
	resp, err := r.client.R().
		Get("https://cn.wordpress.org/wp-admin/tools.php?page=rosetta_manage_release_packages")
	if err != nil {
		return nil, err
	}

	doc, err := htmlquery.Parse(resp.Body)
	if err != nil {
		return nil, err
	}

	node := htmlquery.Find(doc, `/html/body/div[1]/div[2]/div[2]/div[1]/div[2]/table[1]/tbody/tr`)
	if node == nil {
		return nil, fmt.Errorf("未找到版本列表")
	}

	var versions []Version
	for _, n := range node {
		name := htmlquery.FindOne(n, "td")
		if name == nil {
			continue
		}
		// 跳过含有 show-more class 的
		if strings.Contains(htmlquery.SelectAttr(name, "class"), "show-more") {
			continue
		}
		versions = append(versions, Version{
			Name: strings.TrimSpace(htmlquery.InnerText(name)),
		})
	}

	return versions, nil
}

// MergedVersions 合并版本列表
func (r *Tagger) MergedVersions(versions, unMerged []Version) []Version {
	var merged []Version
	for _, v := range versions {
		flag := false
		for _, tv := range unMerged {
			// 已经存在的跳过
			if v.Name == tv.Name {
				flag = true
				break
			}
		}
		if !flag {
			merged = append(merged, v)
		}
	}

	// 过滤一遍，去掉 4.0 以下的远古版本
	var filtered []Version
	for i := 0; i < len(merged); i++ {
		if r.VersionCompare(merged[i].Name, "4.0", "<") {
			continue
		}
		filtered = append(filtered, merged[i])
	}

	return filtered
}

func (r *Tagger) VersionCompare(ver1, ver2, operator string) bool {
	v1 := strings.TrimPrefix(ver1, "v")
	v2 := strings.TrimPrefix(ver2, "v")

	v1s := strings.Split(v1, ".")
	v2s := strings.Split(v2, ".")

	for len(v1s) < len(v2s) {
		v1s = append(v1s, "0")
	}

	for len(v2s) < len(v1s) {
		v2s = append(v2s, "0")
	}

	for i := 0; i < len(v1s); i++ {
		v1i := cast.ToInt(v1s[i])
		v2i := cast.ToInt(v2s[i])

		if v1i > v2i {
			return operator == ">" || operator == ">=" || operator == "!="
		} else if v1i < v2i {
			return operator == "<" || operator == "<=" || operator == "!="
		}
	}
	return operator == "==" || operator == ">=" || operator == "<="
}
