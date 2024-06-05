package client

import "rawh/common"

type Client interface {
	DoRequest(method string, urlString string, customHeaders common.MultiString, data string) error
}
