package config

var ExtConfig Extend

// Extend 扩展配置
//
//	extend:
//	  demo:
//	    name: demo-name
//
// 使用方法： config.ExtConfig......即可！！
type Extend struct {
	AMap    AMap    // 这里配置对应配置文件的结构即可
	AliLive AliLive // 阿里云直播配置
}

type AMap struct {
	Key string
}

type AliLive struct {
	AccessKeyId     string `mapstructure:"accessKeyId"`
	AccessKeySecret string `mapstructure:"accessKeySecret"`
	Region          string `mapstructure:"region"`
	PushDomain      string `mapstructure:"pushDomain"`
	AppName         string `mapstructure:"appName"`
	StreamName      string `mapstructure:"streamName"`
	PushAuthKey     string `mapstructure:"pushAuthKey"`
	PullAuthKey     string `mapstructure:"pullAuthKey"`
	PullDomain      string `mapstructure:"pullDomain"`
}
