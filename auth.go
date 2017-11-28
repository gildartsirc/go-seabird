package seabird

import "github.com/belak/go-plugin"

var auth = plugins.NewRegistry()

func RegisterAuth(name string, factory interface{}) {
    err := plugins.Register(name, factory)
    if err != nil {
        panic(err.Error())
    }
}
