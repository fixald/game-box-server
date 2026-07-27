package alilive

import (
	"crypto/md5"
	"fmt"
	"time"

	"go-admin/config"
)

func InitClient() error {
	conf := config.ExtConfig.AliLive
	if conf.AccessKeyId == "" || conf.AccessKeySecret == "" {
		return fmt.Errorf("阿里云直播配置未设置")
	}
	return nil
}

func generateSign(streamName string, expireTime int64, secret string) string {
	path := fmt.Sprintf("/%s/%s", config.ExtConfig.AliLive.AppName, streamName)
	rand := "0" // "0" by default, other value is ok
	uid := "0"  // "0" by default, other value is ok
	sstring := fmt.Sprintf("%s-%d-%s-%s-%s", path, expireTime, rand, uid, secret)
	auth_key := md5.Sum([]byte(sstring))
	return fmt.Sprintf("%x", auth_key)
}

func GeneratePushURL(streamName string) (string, error) {
	conf := config.ExtConfig.AliLive
	if conf.PushDomain == "" || conf.AppName == "" {
		return "", fmt.Errorf("阿里云直播域名或应用名称未配置")
	}

	expireTime := time.Now().Add(24 * time.Hour).Unix()
	sign := generateSign(streamName, expireTime, conf.PushAuthKey)

	return fmt.Sprintf("rtmp://%s/%s/%s?auth_key=%d-0-0-%s",
		conf.PushDomain, conf.AppName, streamName, expireTime, sign), nil
}

func GeneratePlayURL(streamName string) string {
	conf := config.ExtConfig.AliLive
	expireTime := time.Now().Add(24 * time.Hour).Unix()
	fileName := fmt.Sprintf("%s.flv", streamName)
	sign := generateSign(fileName, expireTime, conf.PullAuthKey)
	return fmt.Sprintf("http://%s/%s/%s?auth_key=%d-0-0-%s", conf.PullDomain, conf.AppName, fileName, expireTime, sign)
}

func GenerateHLSURL(streamName string) string {
	conf := config.ExtConfig.AliLive
	expireTime := time.Now().Add(24 * time.Hour).Unix()
	fileName := fmt.Sprintf("%s.m3u8", streamName)
	sign := generateSign(fileName, expireTime, conf.PullAuthKey)
	return fmt.Sprintf("http://%s/%s/%s?auth_key=%d-0-0-%s", conf.PullDomain, conf.AppName, fileName, expireTime, sign)
}
