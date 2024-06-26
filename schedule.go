package main

import (
	"fmt"
	"github.com/gookit/color"
	"time"
)

func (r *Tagger) KeepLogin() {
	ticker := time.NewTicker(5 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		_, _ = r.client.R().Get("https://cn.wordpress.org/team/")
	}
}

func (r *Tagger) FetchUpdate() {
	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()
	for range ticker.C {
		// 获取版本列表
		versions, err := r.Versions()
		if err != nil {
			color.Redln(err.Error())
			continue
		}
		published, err := r.PublishedVersions()
		if err != nil {
			color.Redln(err.Error())
			continue
		}

		needPublish, err := r.MergedVersions(versions, published)
		if err != nil {
			color.Redln(err.Error())
			continue
		}
		if len(needPublish) == 0 {
			continue
		}

		color.Greenln("需要发布的版本列表: ")
		for _, v := range needPublish {
			color.Greenln(fmt.Sprintf("%s (%s)", v.Name, v.Revision))
		}

		built, err := r.BuiltVersions()
		needBuild, err := r.MergedVersions(needPublish, built)
		if err != nil {
			color.Redln(err.Error())
			continue
		}

		color.Greenln("需要构建的版本列表: ")
		for _, v := range needBuild {
			color.Greenln(fmt.Sprintf("%s (%s)", v.Name, v.Revision))
		}

		for _, v := range needBuild {
			if err = r.Build(v); err != nil {
				color.Redln(err.Error())
			}
		}

		for _, v := range needPublish {
			if err = r.Release(v.Name); err != nil {
				color.Redln(err.Error())
			}
		}
	}
}
