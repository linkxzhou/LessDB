package nginxconf

import (
	"fmt"
	"testing"

	"github.com/linkxzhou/gongx/nginxconf"
)

func Test_NginxConf(t *testing.T) {
	parser := nginxconf.New(
		&nginxconf.ParseOptions{
			SingleFile: true,
		})
	v, err := parser.ParseFile("../test/nginx.conf")
	if err == nil && v != nil {
		for _, v1 := range v {
			fmt.Println("parser nginx.conf: ", v1.String(), ", err: ", err)
		}
	}
}

func Test_Mimetypes(t *testing.T) {
	parser := nginxconf.New(
		&nginxconf.ParseOptions{
			SingleFile: true,
		})
	v, err := parser.ParseFile("../test/mime.types")
	if err == nil && v != nil {
		for _, v1 := range v {
			fmt.Println("parser mime.types: ", v1.String(), ", err: ", err)
		}
	}
}

func Test_ProxyConf(t *testing.T) {
	parser := nginxconf.New(
		&nginxconf.ParseOptions{
			SingleFile: true,
		})
	v, err := parser.ParseFile("../test/proxy.conf")
	if err == nil && v != nil {
		for _, v1 := range v {
			fmt.Println("parser proxy.conf: ", v1.String(), ", err: ", err)
		}
	}
}

func Test_FastcgiConf(t *testing.T) {
	parser := nginxconf.New(
		&nginxconf.ParseOptions{
			SingleFile: true,
		})
	v, err := parser.ParseFile("../test/fastcgi.conf")
	if err == nil && v != nil {
		for _, v1 := range v {
			fmt.Println("parser fastcgi.conf: ", v1.String(), ", err: ", err)
		}
	}
}
